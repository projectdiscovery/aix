package main

import (
	"fmt"

	"github.com/projectdiscovery/aix/internal/runner"
	"github.com/projectdiscovery/gologger"
)

func main() {
	options := &runner.Options{
		Provider:           "openai",
		Model:              "gpt-3.5-turbo", // Specify the model to use
		OpenaiApiKey:       "YOUR_OPENAI_API_KEY",
		Prompt:             "what is the capital of france?",
		Update:             false,
		DisableUpdateCheck: false,
		Output:             "out.txt",
		Jsonl:              false,
		Verbose:            false,
		Silent:             true,
		NoColor:            false,
		Version:            false,
	}

	aixRunner, err := runner.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create runner: %s\n", err)
	}

	result, err := aixRunner.Run()
	if err != nil {
		gologger.Fatal().Msgf("Could not run aix: %s\n", err)
	}

	fmt.Printf("Prompt: %s\n", result.Prompt)
	fmt.Printf("Answer: %s\n", result.Completion)
}