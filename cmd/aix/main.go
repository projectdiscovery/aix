package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/glamour"
	"github.com/logrusorgru/aurora"
	"github.com/projectdiscovery/aix/internal/runner"
	"github.com/projectdiscovery/gologger"
)

func main() {
	options := runner.ParseOptions()
	if options.Stream && options.Jsonl {
		// cannot stream jsonl
		gologger.Fatal().Msgf("Cannot use --stream and --jsonl together")
	}

	var renderer *glamour.TermRenderer

	if !options.NoMarkdown {
		var err error
		renderer, err = glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithEmoji())
		if err != nil {
			gologger.Error().Msgf("Could not create renderer: %s\n", err)
		}
	}

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
		fmt.Printf("[%v] ", aurora.BrightGreen("Completion"))

	}
	if !options.Stream {
		outputData := result.Completion
		if renderer != nil {
			out, err := renderer.Render(outputData)
			if err == nil {
				outputData = out
			}
		}
		fmt.Println(outputData)
	} else {
		// rendering not supported for streaming
		_, _ = io.Copy(os.Stdout, result.CompletionStream)
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
