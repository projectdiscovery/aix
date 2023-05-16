<h1 align="center">
 AIx
<br>
</h1>


<p align="center">
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-_red.svg"></a>
<a href="https://goreportcard.com/badge/github.com/projectdiscovery/aix"><img src="https://goreportcard.com/badge/github.com/projectdiscovery/aix"></a>
<a href="https://pkg.go.dev/github.com/projectdiscovery/aix/pkg/aix"><img src="https://img.shields.io/badge/go-reference-blue"></a>
<a href="https://github.com/projectdiscovery/aix/releases"><img src="https://img.shields.io/github/release/projectdiscovery/aix"></a>
<a href="https://twitter.com/pdiscoveryio"><img src="https://img.shields.io/twitter/follow/pdiscoveryio.svg?logo=twitter"></a>
<a href="https://discord.gg/projectdiscovery"><img src="https://img.shields.io/discord/695645237418131507.svg?logo=discord"></a>
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation">Installation</a> •
  <a href="#help-menu">Usage</a> •
  <a href="#examples">Running AIx</a> •
  <a href="https://discord.gg/projectdiscovery">Join Discord</a>

</p>

<pre align="center">
<b>
	AIx is a cli tool to interact with Large Language Models (LLM) APIs.
</b>
</pre>

![image](https://user-images.githubusercontent.com/8293321/227775051-440d4ed5-f30e-4ec5-bf1d-10310840ab54.png)

## Features
- **AMA with AI** over **CLI**
- **Query LLM APIs** (OpenAI)
- Supports **GPT-3.5 and GPT-4.0** models
- Configurable with OpenAI API key
- Flexible output options

## Installation
To install aix, you need to have Golang 1.19 installed on your system. You can download Golang from [here](https://go.dev/doc/install). After installing Golang, you can use the following command to install aix:


```bash
go install github.com/projectdiscovery/aix/cmd/aix@latest
```

## Prerequisite

> **Note**: Before using aix, make sure to set your [OpenAI API key](https://platform.openai.com/account/api-keys) as an environment variable `OPENAI_API_KEY`.

```bash
export OPENAI_API_KEY=******
````

## Help Menu
You can use the following command to see the available flags and options:

```console
AIx is a cli tool to interact with Large Language Model (LLM) APIs.

Usage:
  ./aix [flags]

Flags:
INPUT:
   -p, -prompt string  prompt to query (input: stdin,string,file)

MODEL:
   -g3, -gpt3        use GPT-3.5 model (default true)
   -g4, -gpt4        use GPT-4.0 model
   -system string[]  system message to send to the model (optional)

CONFIG:
   -ak, -openai-api-key string  openai api key token (input: string,file,env)
   -temperature string          openai model temperature
   -top-p string                openai model top-p

UPDATE:
   -up, -update                 update aix to latest version
   -duc, -disable-update-check  disable automatic aix update check

OUTPUT:
   -o, -output string  file to write output to
   -j, -jsonl          write output in json(line) format
   -v, -verbose        verbose mode
   -silent             display silent output
   -nc, -no-color      disable colors in cli output
   -version            display project version
   -stream             stream output to stdout
   -render             render markdown message returned by the model
```

## Examples

You can use aix to interact with LLM (OpenAI) APIs to query anything and everything in your CLI by specifying the prompts. Here are some examples:

### Example 1: Query LLM with a prompt

```bash
aix -p "What is the capital of France?"
```

### Example 2: Query with GPT-4.0 model
```bash
aix -p "How to install Linux?" -g4
```

### Example 3: Query LLM API with a prompt with STDIN input

```console
echo list top trending web technologies | aix

   ___   _____  __
  / _ | /  _/ |/_/
 / __ |_/ /_>  < 
/_/ |_/___/_/|_|  Powered by OpenAI

   projectdiscovery.io		  

[INF] Current aix version v0.0.1 (latest)
1. Artificial Intelligence (AI) and Machine Learning (ML)
2. Internet of Things (IoT)
3. Progressive Web Apps (PWA)
4. Voice search and virtual assistants
5. Mobile-first design and development
6. Blockchain and distributed ledger technology
7. Augmented Reality (AR) and Virtual Reality (VR)
8. Chatbots and conversational interfaces
9. Serverless architecture and cloud computing
10. Cybersecurity and data protection
11. Mobile wallets and payment gateways
12. Responsive web design and development
13. Social media integration and sharing options
14. Accelerated Mobile Pages (AMP)
15. Content Management Systems (CMS) and static site generators

Note: These technologies are constantly changing and evolving, so this list is subject to change over time.
```

### Example 3: Query LLM API with a prompt and save the output to a file in JSONLine format.
```console
aix -p "What is the capital of France?" -jsonl -o output.txt | jq .

   ___   _____  __
  / _ | /  _/ |/_/
 / __ |_/ /_>  < 
/_/ |_/___/_/|_|  Powered by OpenAI

   projectdiscovery.io		  

[INF] Current aix version v0.0.1 (latest)
{
  "timestamp": "2023-03-26 17:55:42.707436 +0530 IST m=+1.512222751",
  "prompt": "What is the capital of France?",
  "completion": "Paris.",
  "model": "gpt-3.5-turbo"
}
```

### Example 3: Query LLM API in verbose mode
```console
aix -p "What is the capital of France?" -v

   ___   _____  __
  / _ | /  _/ |/_/
 / __ |_/ /_>  < 
/_/ |_/___/_/|_|  Powered by OpenAI

   projectdiscovery.io		  

[INF] Current aix version v0.0.1 (latest)
[VER] [prompt] What is the capital of France?
[VER] [completion] The capital of France is Paris.
```

For more information on the usage of aix, please refer to the help menu with the `aix -h` flag.

## Acknowledgements

- [OpenAI](https://platform.openai.com/docs/introduction) for publishing LLM APIs.
- [sashabaranov](https://github.com/sashabaranov) for building and maintaining [go-openai](https://github.com/sashabaranov/go-openai) library.

--------

<div align="center">

**aix** is made with ❤️ by the [projectdiscovery](https://projectdiscovery.io) team and distributed under [MIT License](LICENSE.md).


<a href="https://discord.gg/projectdiscovery"><img src="https://raw.githubusercontent.com/projectdiscovery/nuclei-burp-plugin/main/static/join-discord.png" width="300" alt="Join Discord"></a>

</div>
