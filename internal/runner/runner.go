package runner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/projectdiscovery/gologger"
	"github.com/sashabaranov/go-openai"
)

// Runner contains the internal logic of the program
type Runner struct {
	options *Options
}

// NewRunner instance
func NewRunner(options *Options) (*Runner, error) {
	return &Runner{
		options: options,
	}, nil
}

// Run the instance
func (r *Runner) Run() error {
	var model string
	if r.options.OpenaiApiKey == "" {
		gologger.Info().Msgf("OPENAI_API_KEY is not configured / provided.")
		os.Exit(1)
	}

	client := openai.NewClient(r.options.OpenaiApiKey)

	if r.options.Gpt3 {
		model = openai.GPT3Dot5Turbo
	}
	if r.options.Gpt4 {
		model = openai.GPT4
	}

	chatGptResp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: r.options.Prompt,
				},
			},
		},
	)
	if err != nil {
		return err
	}

	if len(chatGptResp.Choices) == 0 {
		return fmt.Errorf("no data on response")
	}

	if r.options.Verbose {
		gologger.Verbose().Msgf("[prompt] %s", r.options.Prompt)
		gologger.Verbose().Msgf("[completion] %s", chatGptResp.Choices[0].Message.Content)
		return nil
	}

	if r.options.Jsonl {
		result := Result{
			Timestamp:  time.Now().String(),
			Model:      model,
			Prompt:     r.options.Prompt,
			Completion: chatGptResp.Choices[0].Message.Content,
		}
		gologger.Silent().Msgf("%s", result.JSON())
		return nil
	}
	if r.options.Silent {
		gologger.Silent().Msgf("%s", chatGptResp.Choices[0].Message.Content)
		return nil
	}
	gologger.Info().Msgf("%s", chatGptResp.Choices[0].Message.Content)

	return nil
}
