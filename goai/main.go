package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/franciscoescher/goopenai"
)

func main() {
	filename := os.Args[1]
	completion, err := complete(filename)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(completion)
}

func complete(filename string) (response string, err error) {
	apiKey := os.Getenv("OPENAI_API_TOKEN")
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
	return completions.Choices[len(completions.Choices)-1].Message.Content, err
}
