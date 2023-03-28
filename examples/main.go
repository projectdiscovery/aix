package main

import (
	"fmt"

	"github.com/projectdiscovery/aix/internal/runner"
	"github.com/projectdiscovery/gologger"
)

func main() {
	options := runner.Options{
		OpenaiApiKey:       "API-KEY",
		Prompt:             "what is the capital of france?",
		Gpt3:               true,
		Gpt4:               false,
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

	fmt.Println(result.Prompt)
	fmt.Println(result.Completion)
}
