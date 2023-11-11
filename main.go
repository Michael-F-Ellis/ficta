// A program that monitors one or more files for changes and calls the OpenAI
// completion endpoint with the text  of the file.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/franciscoescher/goopenai"
	"github.com/fsnotify/fsnotify"
)

// TODO #5
// TODO Add @OUT, @/OUT delimiters
const USAGE = `
FICTA v1.2.2

Usage: ficta [options] file1 [file2 ...]

ficta monitors one or more files for changes and calls the OpenAI completion
endpoint with the text of the file. If you pass a filename that doesn't exist,
ficta will create it and write some default content to it.

Options:
   -h Show this help message.
   -b backup extension: the extension for backup files. If -b is not specified,
   ficta will not create backup files when a file is updated.
   -c line comment prefix: the prefix string for comment lines. Default is '//'.
   -y block comment prefix, default = '/*'
   -z block comment suffix, default = '*/'
   Commented lines are excluded from text sent to the OpenAI completion endpoint.

When you save a changed file, ficta will call the OpenAI completion endpoint and
overwrites the file with the original text followed by the completion response,
followed by a one line record containing the model name, max_tokens and
'temperature' and N (number of completions requested). 

A typical model record looks like the following:

AI: gpt-3.5-turbo, 100, 0.700, 1

You may edit the model record with any valid values for model name, max tokens,
temperature, and N; those values will be used for the next completion request.
See the openai.com API documentation to learn more about models, max tokens and
temperature, and N.

You need a valid OpenAI API key and Organization ID to use ficta.  Ficta
expects to find them in environment variables named OPENAI_API_KEY and 
OPENAI_API_ORG.`

var (
	backupExt          string
	lineCommentPrefix  string
	blockCommentPrefix string
	blockCommentSuffix string
)

func main() {
	flag.StringVar(&backupExt, "b", "", "the extension for backup files")
	flag.StringVar(&lineCommentPrefix, "c", "//", "the prefix string for comment lines")
	flag.StringVar(&blockCommentPrefix, "y", "/*", "the prefix string for multi-line comments")
	flag.StringVar(&blockCommentSuffix, "z", "*/", "the suffix string for multi-line comments")
	flag.Usage = func() { fmt.Println(USAGE) }
	flag.Parse()

	files, errors := checkFileArgs(flag.Args())
	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}
	}
	if len(files) == 0 {
		log.Println("No files could be opened")
		fmt.Println(USAGE)
		return
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Please set the OPENAI_API_KEY environment variable")
		fmt.Println(USAGE)
		return
	}
	orgId := os.Getenv("OPENAI_API_ORG")
	if orgId == "" {
		log.Println("Please set the OPENAI_API_ORG environment variable")
		fmt.Println(USAGE)
		return
	}
	// Create a watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("Error:", err)
		return
	}
	defer watcher.Close()

	done := make(chan bool)

	// Enter a goroutine that watches for changes to the files.
	go func() {
		// make a map each filename that is being watched to a bool. We use this to determine if the
		// last file change was done when we appended a response.
		watchedFiles := make(map[string]bool)
		for _, f := range files {
			watchedFiles[f] = false
		}
		log.Printf("Listening for changes to %q", files)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					// We have a write event. The filename is given by event.Name.
					// The watchedFiles map has the names of the files we're watching
					// as keys and a bool indicating if the last file change was done
					// when we appended a response from the API. We want to avoid having
					// those writes cause a send to the API endpoint.
					if v, ok := watchedFiles[event.Name]; ok && v {
						// clear the flag and ignore the write event
						// because it was triggered by our writing the
						// response to the file.
						watchedFiles[event.Name] = false
						continue
					}
					// if we get here, then the last file change was done by the user.
					log.Printf("file changed: %s", event.Name)
					start := time.Now()
					// Call the completion API
					response, err := requestCompletion(event.Name, apiKey, orgId)
					if err != nil {
						log.Println(err)
						continue
					}
					log.Printf("response received: %0.3f elapsed", time.Since(start).Seconds())

					// Rewrite the file with the new content.
					err = overwriteFile(event.Name, backupExt, response)
					if err != nil {
						log.Println(err)
						continue
					}
					// Set a flag so that our write doesn't retrigger change handler.
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

	// Add our filenames to the watcher.
	for _, file := range files {
		err = watcher.Add(file)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	// Wait forever, allowing the goroutine to handle the file changes.
	<-done
}

// checkFileArgs receives a slice of filenames.  Any filenames that don't exist
// are created and a default string supplied as an argument is appended and
// saved. checkFileArgs attempts to open the files and returns a slice of of
// strings containing all the files it was able to open and a slice of errors
// for each file it wasn't able to open.
func checkFileArgs(filenames []string) ([]string, []error) {
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
				// insert default content into file
				err := writeDefaultFileContent(file)
				if err != nil {
					errors = append(errors, err)
					file.Close()
					continue
				}
				defer file.Close()
			}
		} else {
			// file exists, so try to open it
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

// requestCompletion takes a file name and an openai API key and organization id
// and sends the file's content to the OpenAI completion endpoint.  It returns
// the response from the OpenAI completion endpoint and an error if one
// occurred.
func requestCompletion(filename, apiKey, org string) (response string, err error) {
	organization := os.Getenv("OPENAI_API_ORG")

	client := goopenai.NewClient(apiKey, organization)
	text, err := os.ReadFile(filename)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	textstr, aiLine := findLastAILine(string(text))
	cleanText := processAuthorComments(textstr, lineCommentPrefix, blockCommentPrefix, blockCommentSuffix)
	model, max_tokens, temperature, cnt, err := parseAILine(aiLine)
	if err != nil {
		log.Printf("Using default model parameters: Error: %v", err)
	}
	// Escape special characters in text
	escapedText, err := json.Marshal(cleanText)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	r := goopenai.CreateCompletionsRequest{
		Messages: []goopenai.Message{
			{
				Role:    "user",
				Content: string(escapedText),
			},
		},
		Model:       model,
		Temperature: 2 * temperature, // OpenAI API temperature range is 0.0 to 2.0
		MaxTokens:   max_tokens,
		N:           cnt,
	}

	ctx := context.Background()
	completions, err := client.CreateCompletions(ctx, r)
	if err != nil {
		return "", err
	}
	// fmt.Printf("AI: %s, %d, %f, %d\n", model, max_tokens, temperature, cnt)
	// fmt.Printf("Request\n%v\n", r)
	// fmt.Printf("Completions\n%v\n", completions)
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
	// Log the response token counts
	pt := completions.Usage.PromptTokens
	ct := completions.Usage.CompletionTokens
	tt := completions.Usage.TotalTokens
	log.Printf("tokens: prompt=%d, completion=%d, total=%d\n", pt, ct, tt)
	// Create and append model, token limit and temperature as the final line of the response.
	// Remember that temp needs to be scaled down by a factor of 2.
	mdl := r.Model
	maxt := r.MaxTokens
	temp := r.Temperature
	cnt = r.N
	ai := fmt.Sprintf("\n\nAI: %s, %d, %0.3f, %d", mdl, maxt, temp/2, cnt)
	var responses []string
	nChoices := len(completions.Choices)
	switch {
	case nChoices == 1:
		for _, s := range completions.Choices {
			responses = append(responses, s.Message.Content)
		}
	case nChoices > 1:
		for i, s := range completions.Choices {
			// precede each response with a line comment of the from "response n of m"
			responses = append(responses, fmt.Sprintf("%s response %d of %d", lineCommentPrefix, i+1, nChoices))
			responses = append(responses, s.Message.Content)
		}
	default:
		responses = append(responses, completions.Error.Message)
	}
	// catenate the prompt, the response and the AI string. For reasons that
	// aren't yet clear, the responses sometimes contain escape sequences for
	// quotes, tabs and newlines. The unescape function fixes any that are
	// found.
	return textstr + unescape(strings.Join(responses, "\n\n")) + ai, err
}

// findLastAILine returns the AI: line that contains the model, max tokens and
// temperature values to be used for completion.
func findLastAILine(text string) (string, string) {
	lines := strings.Split(text, "\n")

	// Search backwards for an AI: line.
	var aiLine string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.HasPrefix(line, "AI:") {
			aiLine = line
			break
		}
	}
	// Split the text at the beginning of the AI: line and return the two parts.
	// If no AI: line was found, return the entire text as the first part and an
	// empty string as the second part.
	if aiLine != "" {
		index := strings.LastIndex(text, aiLine)
		part1 := text[:index]
		return part1, aiLine
	} else {
		return text, ""
	}
}

// parseAILine parses the AI: line and returns the model, max tokens,
// temperature, and response count values to be used for completion. If there's
// an error parsing the line, it returns a default AI line and an non-nil error.
func parseAILine(line string) (string, int, float64, int, error) {
	defaults := func(err error) (string, int, float64, int, error) {
		return "gpt-3.5-turbo", 100, 0.7, 2, err
	}
	// Split line into fields
	fields := strings.Split(line, ",")
	if len(fields) < 3 || len(fields) > 4 {
		return defaults(fmt.Errorf("Invalid number of fields in line: %q", line))
	}

	// Parse model string field
	modelField := strings.TrimSpace(fields[0])
	if !strings.HasPrefix(modelField, "AI:") {
		return defaults(fmt.Errorf("Invalid string field in line: %q", line))
	}
	model := strings.TrimPrefix(modelField, "AI:")
	model = strings.TrimSpace(model)
	if model == "" {
		return defaults(fmt.Errorf("Empty string field in line: %q", line))
	}

	// Parse max tokens integer field
	maxToksField := strings.TrimSpace(fields[1])
	maxToks, err := strconv.Atoi(maxToksField)
	if err != nil {
		return defaults(fmt.Errorf("Invalid integer field in line: %q", line))
	}
	if maxToks < 0 {
		return defaults(fmt.Errorf("Negative integer field in line: %q", line))
	}

	// Parse float field
	fltField := strings.TrimSpace(fields[2])
	flt, err := strconv.ParseFloat(fltField, 64)
	if err != nil {
		return defaults(fmt.Errorf("Invalid float field in line: %q", line))
	}
	if flt < 0.0 || flt > 1.0 {
		return defaults(fmt.Errorf("Invalid float range in line: %q", line))
	}

	// Parse number of reponses integer field
	var (
		nResp int
	)
	if len(fields) != 4 {
		nResp = 1
	} else {
		nRespField := strings.TrimSpace(fields[3])
		nResp, err = strconv.Atoi(nRespField)
		if err != nil {
			return defaults(fmt.Errorf("Invalid integer field in line: %q", line))
		}
		if nResp <= 0 {
			return defaults(fmt.Errorf("Zero or negative integer field in line: %q", line))
		}
	}

	// Return parsed fields
	return model, maxToks, flt, nResp, nil
}

// unescape unescapes a string, replacing backslash escaped characters with
// the corresponding unescaped runes.
func unescape(input string) string {
	var buf bytes.Buffer
	escaped := false

	for _, r := range input {
		if escaped {
			switch r {
			case 't':
				buf.WriteRune('\t')
			case 'n':
				buf.WriteRune('\n')
			case '\\':
				buf.WriteRune('\\')
			case '"':
				buf.WriteRune('"')
			default:
				// Write the backslash and the current rune
				buf.WriteRune('\\')
				buf.WriteRune(r)
			}
			escaped = false
		} else if r == '\\' {
			escaped = true
		} else {
			buf.WriteRune(r)
		}
	}

	if escaped {
		// Write the trailing backslash
		buf.WriteRune('\\')
	}

	return buf.String()
}

// writeDefaultFileContent is called to put some initial content into
// files that ficta creates.
func writeDefaultFileContent(file *os.File) error {
	const defaultContent = `Continue the story that starts below.

Once upon a time there were three weasels named Willy, Worgus and Wishbone. One bright spring morning, Willy said to Worgus, "Hey, dude, what's for breakfast?"

AI: gpt-3.5-turbo, 100, 0.700, 1`
	_, err := file.WriteString(defaultContent)
	return err
}

// overwriteFile rewrites a file with new content. If backExt is not "", it
// creates a backup of the original file with the given extension.
func overwriteFile(filename, bakExt, newContent string) error {
	// create backup file
	if bakExt != "" {
		backupFilename := replaceExtension(filename, bakExt)
		if err := copyFile(filename, backupFilename); err != nil {
			return err
		}
	}
	// overwrite original file with content
	if err := os.WriteFile(filename, []byte(newContent), 0644); err != nil {
		return err
	}
	return nil
}

func copyFile(src string, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}

// replaceExtension takes in a filename and a new extension as input parameters
// and returns the filename with the updated extension. If the newExt parameter
// is "", it returns the filename.
func replaceExtension(filename, newExt string) string {
	// Check if filename includes extension
	ext := filepath.Ext(filename)
	if ext == "" {
		// If no extension, simply append the new extension
		if newExt == "" {
			// Don't add a "."
			return filename
		} else {
			return filename + "." + newExt
		}
	}
	// Remove the existing extension and replace with new extension
	fname := strings.TrimSuffix(filename, ext)
	if newExt == "" {
		return fname
	} else {
		return fname + "." + strings.TrimPrefix(newExt, ".")
	}
}

// processAuthorComments removes the author comments from a text string.
// The arguments lcprefix, bcprefix, and lcprefix2 are comment delimiters
// for line and block comments.
func processAuthorComments(text, lcprefix, bcprefix, bcsuffix string) string {
	var (
		include = true
	)
	lines := strings.Split(text, "\n")
	stripped := []string{} // accumulator for lines that aren't commented out
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, lcprefix) {
			continue // reject line comments
		}
		if strings.HasSuffix(trimmed, bcsuffix) {
			include = true
			continue
		}
		if strings.HasPrefix(trimmed, bcprefix) {
			include = false
			continue
		}
		// Line is not commented out, so add it to the accumulator
		if include {
			stripped = append(stripped, line)
		}
	}
	return strings.Join(stripped, "\n")
}
