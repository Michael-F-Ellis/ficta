// A program that monitors one or more files for changes and calls the OpenAI completion endpoint with the text  of the file.
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	files, errors := openFiles()
	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}
	}
	if len(files) == 0 {
		log.Println("No files could be opened")
		return
	}
	for _, file := range files {
		defer file.Close()
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Please set the OPENAI_API_KEY environment variable")
		return
	}
	// set up a timer to check for changes in the file(s) every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	for {
		select {
		case <-ticker.C:
			for _, f := range files {
				// read the contents of the file
				bytes, err := ioutil.ReadAll(f)
				if err != nil {
					log.Fatal(err)
				}
				text := string(bytes)

				// call the OpenAI completion endpoint with the text of the file
				completionURL := "https://api.openai.com/v1/completions"
				reqBody := fmt.Sprintf(`{"prompt": "%s", "max_tokens": 1000}`, text)
				req, err := http.NewRequest("POST", completionURL, strings.NewReader(reqBody))
				if err != nil {
					log.Fatal(err)
				}
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					log.Fatal(err)
				}
				defer resp.Body.Close()

				// print the response from the OpenAI API
				respBytes, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(respBytes))
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

// getFilenames reads filenames from the command line.
func getFilenames() []string {
	var filenames []string
	filenames = append(filenames, os.Args[1:]...)
	return filenames
}

// openFiles reads one or more filenames from the command line. Any filenames that don't exist are created
// and a default string supplied as an argument is appended and saved. openFiles attempts to open the files
// and returns a slice of *os.File containing all the files it was able to open and a slice of errors for each file
// it wasn't able to open.
func openFiles() ([]*os.File, []error) {
	filenames := getFilenames()
	var files []*os.File
	var errors []error
	for _, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		files = append(files, file)
	}
	return files, errors
}
