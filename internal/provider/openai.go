package provider

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider communicates with the OpenAI API.
type OpenAIProvider struct {
	client *openai.Client
	model  string
}

// NewOpenAI creates a new OpenAIProvider.
func NewOpenAI(apiKey string, model string) (*OpenAIProvider, error) {
	if apiKey == "" {
		return nil, errors.New("OpenAI API key is not configured")
	}
	if model == "" {
		model = openai.GPT3Dot5Turbo // Default model if not specified
	}
	return &OpenAIProvider{
		client: openai.NewClient(apiKey),
		model:  model, // Store the selected model
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	models, err := p.client.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	var modelIDs []string
	for _, model := range models.Models {
		modelIDs = append(modelIDs, model.ID)
	}
	return modelIDs, nil
}

func (p *OpenAIProvider) Chat(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (Result, error) {
	req := p.createChatCompletionRequest(prompt, systemPrompt, false, opts)
	resp, err := p.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return Result{}, err
	}
	if len(resp.Choices) == 0 {
		return Result{}, errors.New("no choices in response")
	}
	return Result{
		Completion: resp.Choices[0].Message.Content,
		Model:      resp.Model,
	}, nil
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (StreamResult, error) {
	req := p.createChatCompletionRequest(prompt, systemPrompt, true, opts)
	stream, err := p.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return StreamResult{}, err
	}

	r, w := io.Pipe()
	errChan := make(chan error, 1)

	go func() {
		defer func() { _ = stream.Close() }()
		defer func() { _ = w.Close() }()
		defer close(errChan)

		for {
			response, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				errChan <- err
				return
			}
			if len(response.Choices) > 0 {
				_, writeErr := w.Write([]byte(response.Choices[0].Delta.Content))
				if writeErr != nil {
					errChan <- writeErr
					return
				}
			}
		}
	}()

	return StreamResult{
		CompletionStream: r,
		Error:            errChan,
	}, nil
}

func (p *OpenAIProvider) createChatCompletionRequest(prompt string, systemPrompt string, stream bool, opts ChatOptions) openai.ChatCompletionRequest {
	messages := []openai.ChatCompletionMessage{}
	if systemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: prompt,
	})

	req := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: messages,
		Stream:   stream,
	}

	if opts.Temperature != 0 {
		req.Temperature = opts.Temperature
	}
	if opts.TopP != 0 {
		req.TopP = opts.TopP
	}

	return req
}
