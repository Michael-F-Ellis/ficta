# ficta

Usage: ficta file1 [file2 ...]

ficta monitors one or more files for changes and calls the OpenAI completion
endpoint with the text of the file. If you pass a filename that doesn't exist,
ficta will create it and write some default content to it.

When you save a changed file, ficta will call the OpenAI completion endpoint
with the original text followed by the completion response, followed by a one
line record containing the model name, max_tokens and 'temperature' settings
passed with the completion request.

A typical model record looks like the following:

`AI: gpt-3.5-turbo, 400, 0.700`

You may edit the model record with any valid values for model name, max tokens
and temperature and those values will be used for the next completion request.
See the openai.com API documentation to learn more about models, max tokens and
temperature.

You need a valid OpenAI API key and Organization ID to use ficta.  Ficta
expects to find them in environment variables named `OPENAI_API_KEY` and 
`OPENAI_API_ORG`.
