Continue the README file that starts below.
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


ficta is a a command line program that lets you OpenAI's completion API 
from any text editor. ## Installation


## Usage

To use `ficta`, simply run the command followed by the names of the files you wish to monitor:

```bash
ficta file1.txt file2.txt
```

If a file doesn't exist, `ficta` will create it and add the following default content to it.
```
Continue the story that starts below.

Once upon a time there were three weasels named Willy, Worgus and Wishbone. One bright spring morning, Willy said to Worgus, "Hey, dude, what's for breakfast?"

AI: gpt-3.5-turbo, 400, 0.700
```
Once you have started monitoring a file, any changes you make to it will trigger a call to the OpenAI completion endpoint. The original text of the file will be sent to the endpoint, along with any settings you have specified (such as model name, max tokens, and temperature). 

The completion response will be appended to the original text, and the resulting text will be saved back to the file. A record of the model name, max tokens, and temperature settings used for the completion request will also be included in the file.

You can edit the model record in the file to change the settings that will be used for the next completion request.

## API Key and Organization ID

To use `ficta`, you will need a valid OpenAI API key and Organization ID. These should be stored in environment variables named `OPENAI_API_KEY` and `OPENAI_API_ORG`, respectively.

If you do not have an OpenAI API key, you can sign up for one on the OpenAI website.

Here's a usage example starting from the default content for a file that ficta creates for you. To keep things brief, let's say you change the max tokens value at the bottom of the file to 100 and save the file. A few seconds later, the file will be updated to something similar the following:


*Continue the story that starts below.*

*Once upon a time there were three weasels named Willy, Worgus and Wishbone. One bright spring morning, Willy said to Worgus, "Hey, dude, whats's for breakfast?"*

*Worgus scratched his head and replied, "I don't know, Willy. We're running low on food and we can't keep relying on Wishbone to catch all the rabbits."*

*Wishbone, who had been napping in the sun, suddenly perked up at the sound of his name. "What's the problem, guys?" he asked.*

*"We're hungry and we need to find some food," explained Willy.*

*AI: gpt-3.5-turbo, 100, 0.700*

Assuming you're writing a children's story, you might decide you like the flow of what *gpt* generated but figure that a story about weasels eating rabbits and calling each other 'dude' isn't very marketable. So you decide to make it about rabbits and vegetables.

After a few quick edits and a save, you get


*Continue the story that starts below.*

*Once upon a time there were three rabbits  named Willy, Worgus and Wishbone. One bright midsummer morning, Willy said to Worgus, "Hey, Worgus, whats's for breakfast?"*

*Worgus scratched his head and replied, "I don't know, Willy. We're running low on food and we can't keep relying on Wishbone to grow all the vegetables."*

*Wishbone, who had been napping in the sun, suddenly perked up at the sound of his name. "What's the problem, guys?" he asked.*

*"We're hungry and we need to find some food," explained Willy.*

*Wishbone yawned and stretched. "No problem," he said. "I know where we can find some delicious carrots and lettuce. Follow me!"*

*The three rabbits scampered through the fields and over the hills until they came to a lush garden filled with rows of vegetables. Wishbone led the way, his nose twitching with excitement as he sniffed out the freshest produce.*

*"Look at all these carrots!" exclaimed Willy, his mouth watering. "And the lettuce is so crisp and green!"*

*AI: gpt-3.5-turbo, 100, 0.700*



