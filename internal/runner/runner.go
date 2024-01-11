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
		var buff bytes.Buffer
		for _, model := range models.Models {
			buff.WriteString(fmt.Sprintf("%s\n", model.ID))
		}

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
