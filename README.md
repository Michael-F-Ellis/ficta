# ficta
**NEW:** `ficta` now supports using local LLM server endpoints that mimic the OpenAI API `v1/chat/completions` endpoint, e.g. `llama.cpp server`

`ficta` is a command line program that lets you use OpenAI's completion API from any text editor.

`ficta` exists because I found it frustrating to write short stories and essays via the ChatGPT web interface. With `ficta`, the developing story becomes the prompt. `ficta` also lets you change LLM model and parameters freely in mid-stream.

`ficta` attempts to adhere to the Unix/Linux philosophy that programs should do one thing well and cooperate with other programs. In the case of `ficta`, you specify some text files to watch and it monitors them for changes. When you edit and save a file, `ficta` handles sending the contents of the file to the OpenAI API endpoint and updating your file with the response.

## Installation
`ficta` is written in Go. It could easily have been written in Python. I chose Go for its superior error handling, networking, and co-routines.

Go is a compiled language, so you need to have the Go toolchain installed on the machine you use to compile `ficta`. The compilation produces a single executable binary you can run on any machine having the same architecture and OS.

To install Go, see https://go.dev/doc/install

Then clone this repo and build `ficta` with

```bash
go build
```

Copy the `ficta` binary somewhere in your `$PATH`, (or run `go install`) and you're ready to run. But see [the API Key section](#api-key-and-organization-id) to learn about two environment variables `ficta` needs to authenticate with OpenAI.

## Usage

To use `ficta`, simply run the command followed by the names of the files you wish to monitor:

```bash
ficta [options] file1.txt file2.txt ... fileN.txt
```
Valid options are:
```
   -h Show this help message.
   -b backupExtension: the extension for backup files. If -b is not specified, ficta will not create backup files when a file is updated.
   -c line comment prefix, default = //
   -y block comment prefix, default = /*
   -z block comment suffix, default = */
```
If you supply a filename that doesn't exist, `ficta` will create it and initialize it with some default content.

Once you have started monitoring a file, any changes you make to it will trigger a call to the OpenAI `v1/chat/completion` endpoint. The original text of the file will be sent to the endpoint, along with any settings you have specified (such as model name, max tokens, and temperature). 

The completion response will be appended to the original text, and the resulting text will be saved back to the file. A record of the model name, max tokens, temperature and number of completions settings used for the completion request will also be included in the file. You can edit the model record to adjust the settings to your needs on subsequent completion requests.

### Author Comments
You may want to exclude some lines of the text from the AI input, either as notes to yourself or in order to keep the input smaller than the maximum input to the AI model you are using.

 `Ficta` supports line and block comments. By default, the comment delimiters are the familiar `//`, `/*`, and `*/` used in C++, Go, and similar programming languages, but you can change them with command line options when you start `ficta`.

 The default delimiters have the advantage of making it easier to adapt existing syntax hightlighting rules to help you distinguish comments from input text. The `ficta` repository includes a `vscode` extension named `AIT` that detects and highlights comments. You'll need to manually copy the folder to your vscode extensions directory and use the file extension `.ait` on your input files.
## API Key and Organization ID

To use `ficta`, you will need a valid OpenAI API key and Organization ID. These should be stored in environment variables named `OPENAI_API_KEY` and `OPENAI_API_ORG`, respectively.

If you do not have an OpenAI API key, you can sign up for one on the OpenAI website.

## Usage example
When `ficta` creates a new text file for you, it initializes it with the following default content. We'll use that content to illustrate development of a fiction story. Here's the initial content.

----
*Continue the story that starts below.*

*Once upon a time there were three weasels named Willy, Worgus and Wishbone. One bright spring morning, Willy said to Worgus, "Hey, dude, what's for breakfast?"*

*AI: gpt-3.5-turbo, 400, 0.700*

----
The default content has three lines of text:
 1. A brief ***prompt*** that tells the AI we're writing a story. You can do without this sometimes if you start with enough of the story, but adding the initial prompt is more reliable. You can also add instructions to the prompt to influence the LLM's writing style. For instance, I often add something like *"Prefer dialog to narrative. Use sights, sounds, sensations, gestures, facial expressions and involuntary actions to convey emotions."*

 2. The ***text*** of the story so far. In this case, a single opening sentence.
 
 3. The ***"AI:"*** line that tells `ficta` which LLM model to use, the maximum number of 'tokens' to generate, and the 'temperature'. You can specify any valid OpenAI model, e.g. `gpt-4` or you can specify `url` to tell `ficta` to use an alternate endpoint. You must specify the alternate endpoint with the -u parameter when you launch ficta for this to work.
  
   A `token` is a short sequence of characters. Typically, 100 tokens is about 75 words. `Temperature` is a parameter that governs the extent to which the LLM will randomly deviate from the next most likely word as it generates text. Temperature must be in the range 0.0 to 1.0. with 0 meaning little or no deviation and 1.0 meaning the AI will be more "creative"


To keep things brief, let's say you change the max tokens value at the bottom of the file to 100 and save the file. A few seconds later, the file will be updated to something similar to the following. I've formatted the input text in bold so you can see what was added. Notice that our change to the AI line was preserved in the output. This allows you to continue with the same model parameters or edit them as needed.

----
**Continue the story that starts below.**

**Once upon a time there were three weasels named Willy, Worgus and Wishbone. One bright spring morning, Willy said to Worgus, "Hey, dude, whats's for breakfast?"**

*Worgus scratched his head and replied, "I don't know, Willy. We're running low on food and we can't keep relying on Wishbone to catch all the rabbits."*

*Wishbone, who had been napping in the sun, suddenly perked up at the sound of his name. "What's the problem, guys?" he asked.*

*"We're hungry and we need to find some food," explained Willy.*

**AI: gpt-3.5-turbo, 100, 0.700**

----
Assuming you're writing a children's story, you might decide you like the flow of what *gpt* generated but figure that a story about weasels eating rabbits and calling each other 'dude' isn't very marketable. So you decide to make it about rabbits and vegetables.

After a few quick edits and a save, you get

----
**Continue the story that starts below.**

**Once upon a time there were three `rabbits`  named Willy, Worgus and Wishbone. One bright `midsummer` morning, Willy said to Worgus, "Hey, `Worgus`, whats's for breakfast?"**

**Worgus scratched his head and replied, "I don't know, Willy. We're running low on food and we can't keep relying on Wishbone to `grow all the vegetable`s."**

**Wishbone, who had been napping in the sun, suddenly perked up at the sound of his name. "What's the problem, guys?" he asked.**

**"We're hungry and we need to find some food," explained Willy.**

*Wishbone yawned and stretched. "No problem," he said. "I know where we can find some delicious carrots and lettuce. Follow me!"*

*The three rabbits scampered through the fields and over the hills until they came to a lush garden filled with rows of vegetables. Wishbone led the way, his nose twitching with excitement as he sniffed out the freshest produce.*

*"Look at all these carrots!" exclaimed Willy, his mouth watering. "And the lettuce is so crisp and green!"*

*AI: gpt-3.5-turbo, 100, 0.700*

----
So maybe at this point you want to control the direction of the story by, say, having the farmer's wife shoo the hungry rabbits away by adding a line like,

**Mrs. McGreedy, the farmer's wife, came out the house running and shouting, "Get away from my garden, rabbits! Those vegetables are mine!"**

Save again, and the LLM will pick up the cue, continuing the story. If you don't like the result, you can always delete the continuation and try again.

----
  *The three rabbits froze in fear as Mrs. McGreedy approached with a broom in her hand. But Wishbone, being the bravest of the group, stepped forward and spoke up.*

  *"Please don't be angry, Mrs. McGreedy," he said. "We were just so hungry and we couldn't find any food anywhere else. We'll leave right away and never come back."*

  *Mrs. McGreedy softened at the sight of the small, innocent rabbits and put down her broom. She sighed and said, "Alright, I understand you're hungry. But you can't just come and take my vegetables without asking. How about I give you some carrots and lettuce in exchange for some help around the farm?"*

  *The rabbits looked at each other, surprised and grateful for Mrs. McGreedy's offer. "What kind of help?" asked Worgus.*

*AI: gpt-3.5-turbo, 100, 0.700*

----

And so on until the text seems complete and ready for editing.

To stop monitoring files, kill `ficta` from the terminal window where you launched it. While `ficta` is running, you may find it useful to leave the terminal window open. `ficta` logs status messages and errors to stdout as shown below:

```bash
% ./ficta weasels.txt
2023/05/09 17:42:50 Listening for changes to ["weasels.txt"]
2023/05/09 17:43:00 file changed: weasels.txt
2023/05/09 17:43:06 tokens: prompt=315, completion=100, total=415
2023/05/09 17:43:06 response received: 6.775 elapsed
2023/05/09 17:43:38 file changed: weasels.txt
2023/05/09 17:43:44 tokens: prompt=315, completion=100, total=415
2023/05/09 17:43:44 response received: 5.804 elapsed
2023/05/09 17:44:14 file changed: weasels.txt
2023/05/09 17:44:20 tokens: prompt=423, completion=100, total=523
2023/05/09 17:44:20 response received: 5.569 elapsed
2023/05/09 17:45:52 file changed: weasels.txt
2023/05/09 17:45:58 tokens: prompt=505, completion=100, total=605
2023/05/09 17:45:58 response received: 6.401 elapsed
```

## Acknowledgments
`ficta` uses Francisco Escher's excellent [goopenai](github.com/franciscoescher/goopenai) package to interface with the OpenAI API.
