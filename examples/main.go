package main

import (
	"fmt"

	"github.com/projectdiscovery/aix/internal/runner"
	"github.com/projectdiscovery/gologger"
)

func main() {
	options := runner.ParseOptions()
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
