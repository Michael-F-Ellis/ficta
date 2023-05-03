// A program that monitors one or more files for changes and calls the OpenAI completion endpoint with the text  of the file.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/franciscoescher/goopenai"
	"github.com/fsnotify/fsnotify"
)

func main() {
	start := time.Now()
	files, errors := checkFiles()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}
	}
	if len(files) == 0 {
		log.Println("No files could be opened")
		return
	}
	apiKey := os.Getenv("OPENAI_API_TOKEN")
	if apiKey == "" {
		log.Println("Please set the OPENAI_API_TOKEN environment variable")
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)
	log.Printf("%0.3f elapsed", time.Since(start).Seconds())
	go func() {
		// make a map each filename that is being watched to a bool. We use this to determine if the
		// last file change was done when we appended a response.
		watchedFiles := make(map[string]bool)
		for _, f := range files {
			watchedFiles[f] = false
		}
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					if v, ok := watchedFiles[event.Name]; ok && v {
						watchedFiles[event.Name] = false
						log.Printf("%s has been modified by us\n", event.Name)
						continue
					}
					log.Printf("file changed: %0.3f elapsed", time.Since(start).Seconds())
					response, err := complete(event.Name, apiKey)
					if err != nil {
						log.Println(err)
						continue
					}
					log.Printf("response received: %0.3f elapsed", time.Since(start).Seconds())
					unquoted, err := strconv.Unquote(response)
					if err != nil {
						log.Println(err)
						unquoted, err = strconv.Unquote(fmt.Sprintf("`%s`", response))
						if err != nil {
							log.Println(err)
						}
					}

					// append the response to the file
					file, err := os.OpenFile(event.Name, os.O_APPEND|os.O_WRONLY, 0644)
					if err != nil {
						log.Println(err)
						continue
					}

					_, err = file.WriteString(unquoted)
					if err != nil {
						log.Println(err)
						file.Close()
						continue
					}
					file.Close()
					watchedFiles[event.Name] = true
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Error:", err)
			}
		}
	}()

	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}

	<-done
}

// getFilenames reads filenames from the command line.
func getFilenames() []string {
	var filenames []string
	filenames = append(filenames, os.Args[1:]...)
	return filenames
}

// checkFiles reads one or more filenames from the command line. Any filenames that don't exist are created
// and a default string supplied as an argument is appended and saved. checkFiles attempts to open the files
// and returns a slice of *os.File containing all the files it was able to open and a slice of errors for each file
// it wasn't able to open.
func checkFiles() ([]string, []error) {
	filenames := getFilenames()
	var goodfiles []string
	var errors []error
	for _, filename := range filenames {
		// if filename doesn't exist, create it
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			file, err := os.Create(filename)
			if err != nil {
				errors = append(errors, err)
			} else {
				// make file read/write
				err = file.Chmod(0644)
				if err != nil {
					errors = append(errors, err)
					file.Close()
					continue
				}
				goodfiles = append(goodfiles, filename)
				defer file.Close()
			}
		} else {
			file, err := os.Open(filename)
			if err != nil {
				errors = append(errors, err)
			} else {
				// make file read/write
				err = file.Chmod(0644)
				if err != nil {
					errors = append(errors, err)
					file.Close()
					continue
				}
				goodfiles = append(goodfiles, filename)
				defer file.Close()
			}
		}
	}

	return goodfiles, errors
}

// complete takes a file name and an openai API key and sends it to the OpenAI completion endpoint.
// It returns the response from the OpenAI completion endpoint and an error if one occurred.
func complete(filename string, apiKey string) (response string, err error) {
	organization := os.Getenv("OPENAI_API_ORG")

	client := goopenai.NewClient(apiKey, organization)
	text, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error:", err)
		return
	}

	// Escape special characters in text
	escapedText, err := json.Marshal(string(text))
	if err != nil {
		log.Println("Error:", err)
		return
	}
	r := goopenai.CreateCompletionsRequest{
		Model: "gpt-3.5-turbo",
		Messages: []goopenai.Message{
			{
				Role:    "user",
				Content: string(escapedText),
			},
		},
		Temperature: 0.7,
	}
	ctx := context.Background()
	completions, err := client.CreateCompletions(ctx, r)
	if err != nil {
		panic(err)
	}

	/* Response should be like this
	{
	  "id": "chatcmpl-xxx",
	  "object": "chat.completion",
	  "created": 1678667132,
	  "model": "gpt-3.5-turbo-0301",
	  "usage": {
	    "prompt_tokens": 13,
	    "completion_tokens": 7,
	    "total_tokens": 20
	  },
	  "choices": [
	    {
	      "message": {
	        "role": "assistant",
	        "content": "\n\nThis is a test!"
	      },
	      "finish_reason": "stop",
	      "index": 0
	    }
	  ]
	}
	*/
	pt := completions.Usage.PromptTokens
	ct := completions.Usage.CompletionTokens
	tt := completions.Usage.TotalTokens
	log.Printf("tokens: prompt=%d, completion=%d, total=%d\n", pt, ct, tt)
	return completions.Choices[len(completions.Choices)-1].Message.Content, err
}
