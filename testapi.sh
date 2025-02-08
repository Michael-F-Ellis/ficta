#!/bin/bash

# Cody API endpoint
CODY_URL="https://sourcegraph.com/api/completions/cody"

# Your Sourcegraph access token
TOKEN=$CODYTOKEN
echo $TOKEN

# Sample completion request with verbose output
curl -v -X POST "${CODY_URL}" \
  -H "Authorization: token ${TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {
        "role": "user",
        "content": "Write a Go function that reverses a string"
      }
    ],
    "model": "anthropic/claude-2.0",
    "temperature": 0.7,
    "max_tokens": 100,
    "n": 1
  }'