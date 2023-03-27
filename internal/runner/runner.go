package runner

import (
	"context"
	"fmt"
	"time"

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
func (r *Runner) Run() (*Result, error) {
	var model string
	if r.options.OpenaiApiKey == "" {
		return &Result{}, fmt.Errorf("OPENAI_API_KEY is not configured / provided.")
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
		return &Result{}, err
	}

	if len(chatGptResp.Choices) == 0 {
		return &Result{}, fmt.Errorf("no data on response")
	}

	result := &Result{
		Timestamp:  time.Now().String(),
		Model:      model,
		Prompt:     r.options.Prompt,
		Completion: chatGptResp.Choices[0].Message.Content,
	}

	return result, nil
}
