package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/projectdiscovery/gologger"
	"github.com/projectdiscovery/manx/internal/runner"
)

func main() {
	options := runner.ParseOptions()
	manxRunner, err := runner.NewRunner(options)
	if err != nil {
		gologger.Fatal().Msgf("Could not create runner: %s\n", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// Setup close handler
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal, Exiting...")
		os.Exit(0)
	}()

	result, err := manxRunner.Run()
	if err != nil {
		gologger.Fatal().Msgf("Could not run manx: %s\n", err)
	}

	if options.Verbose {
		gologger.Verbose().Msgf("[prompt] %s", result.Prompt)
		gologger.Verbose().Msgf("[completion] %s", result.Completion)
	} else if options.Jsonl {
		gologger.Silent().Msgf("%s", result.JSON())
	} else if options.Silent {
		gologger.Silent().Msgf("%s", result.Completion)
	} else {
		gologger.Info().Msgf("%s", result.Completion)
	}
}
