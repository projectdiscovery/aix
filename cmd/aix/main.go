package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/logrusorgru/aurora"
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

	if options.Jsonl {
		gologger.DefaultLogger.Print().Msg(result.JSON())
		if options.Output != "" {
			if err := os.WriteFile(options.Output, []byte(result.JSON()), 0644); err != nil {
				gologger.Error().Msgf("failed to save output to file %v got %v", options.Output, err)
			}
			return
		}
	} else if options.Verbose {
		aurora := aurora.NewAurora(!options.NoColor)
		gologger.DefaultLogger.Print().Msgf("[%v] %v", aurora.BrightYellow("Prompt"), result.Prompt)
		gologger.DefaultLogger.Print().Msgf("[%v] %v", aurora.BrightGreen("Completion"), result.Completion)
	} else {
		gologger.DefaultLogger.Print().Msg(result.Completion)
	}
	if options.Output != "" {
		if err := os.WriteFile(options.Output, []byte(result.Completion), 0644); err != nil {
			gologger.Error().Msgf("failed to save output to file %v got %v", options.Output, err)
		}
	}
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// Setup close handler
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal, Exiting...")
		os.Exit(0)
	}()
}
