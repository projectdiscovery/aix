package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	errorutil "github.com/projectdiscovery/utils/errors"
	"github.com/sashabaranov/go-openai"
)

// ErrNoKey is returned when no key is provided
var ErrNoKey = errorutil.New("OPENAI_API_KEY is not configured / provided.")

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
		return &Result{}, ErrNoKey
	}

	client := openai.NewClient(r.options.OpenaiApiKey)

	if r.options.Gpt3 {
		model = openai.GPT3Dot5Turbo
	}
	if r.options.Gpt4 {
		// use turbo preview by default
		model = openai.GPT4TurboPreview
	}
	if r.options.Model != "" {
		model = r.options.Model
	}

	if r.options.ListModels {
		models, err := client.ListModels(context.Background())
		if err != nil {
			return &Result{}, err
		}

		// Categorize models into gpt, o1, and others
		var gptModels, o1Models, otherModels []string
		for _, model := range models.Models {
			switch {
			case strings.HasPrefix(model.ID, "gpt") || strings.HasPrefix(model.ID, "chatgpt"):
				gptModels = append(gptModels, model.ID)
			case strings.HasPrefix(model.ID, "o1"):
				o1Models = append(o1Models, model.ID)
			default:
				otherModels = append(otherModels, model.ID)
			}
		}

		// Use a buffer to build the output
		var buff bytes.Buffer

		// Print GPT models
		buff.WriteString("## GPT Models:\n\n")
		printModelsInGrid(&buff, gptModels, 2) // Print in 2 columns
		buff.WriteString("\n")

		// Print O1 models
		buff.WriteString("## O1 Models:\n\n")
		printModelsInGrid(&buff, o1Models, 2) // Print in 2 columns
		buff.WriteString("\n")

		// Print Other models
		buff.WriteString("## Other Models:\n\n")
		printModelsInGrid(&buff, otherModels, 2) // Print in 2 columns
		buff.WriteString("\n")

		result := &Result{
			Timestamp: time.Now().String(),
			Model:     model,
			Prompt:    r.options.Prompt,
		}

		if r.options.Stream {
			result.SetupStreaming()
			go func(res *Result) {
				defer res.CloseCompletionStream()
				res.WriteCompletionStreamResponse(buff.String())
			}(result)
		} else {
			result.Completion = buff.String()
		}
		return result, nil
	}

	chatReq := openai.ChatCompletionRequest{
		Model:    model,
		Messages: []openai.ChatCompletionMessage{},
	}

	if len(r.options.System) != 0 {
		chatReq.Messages = append(chatReq.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: strings.Join(r.options.System, "\n"),
		})
	}

	if r.options.Prompt != "" {
		chatReq.Messages = append(chatReq.Messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: r.options.Prompt,
		})
	}

	if len(chatReq.Messages) == 0 {
		return &Result{}, fmt.Errorf("no prompt provided")
	}

	if r.options.Temperature != 0 {
		chatReq.Temperature = r.options.Temperature
	}
	if r.options.TopP != 0 {
		chatReq.TopP = r.options.TopP
	}

	result := &Result{
		Timestamp: time.Now().String(),
		Model:     model,
		Prompt:    r.options.Prompt,
	}

	switch {
	case r.options.Stream:
		// stream response
		result.SetupStreaming()
		go func(res *Result) {
			defer res.CloseCompletionStream()
			chatReq.Stream = true
			stream, err := client.CreateChatCompletionStream(context.TODO(), chatReq)
			if err != nil {
				res.Error = err
				return
			}
			for {
				response, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					res.Error = err
					return
				}
				if len(response.Choices) == 0 {
					res.Error = fmt.Errorf("got empty response")
					return
				}
				res.WriteCompletionStreamResponse(response.Choices[0].Delta.Content)
			}
		}(result)
	default:
		chatGptResp, err := client.CreateChatCompletion(context.TODO(), chatReq)
		if err != nil {
			return &Result{Error: err}, err
		}
		if len(chatGptResp.Choices) == 0 {
			return &Result{}, fmt.Errorf("no data on response")
		}
		result.Completion = chatGptResp.Choices[0].Message.Content
	}

	return result, nil
}

// printModelsInGrid prints models in a grid layout with a specified number of columns
func printModelsInGrid(buff *bytes.Buffer, models []string, columns int) {
	// Calculate the maximum length of model names in the list
	maxLength := 0
	for _, model := range models {
		if len(model) > maxLength {
			maxLength = len(model)
		}
	}

	columnWidth := maxLength + 5

	// Print models in a grid
	for i, model := range models {
		buff.WriteString(fmt.Sprintf("%-*s", columnWidth, model))
		// Move to the next line after every `columns` models
		if (i+1)%columns == 0 {
			buff.WriteString("\n")
		}
	}
	// Ensure the last line ends properly
	if len(models)%columns != 0 {
		buff.WriteString("\n")
	}
}