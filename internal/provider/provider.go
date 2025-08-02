package provider

import (
	"context"
	"io"
)

// Result holds the common result from a Chat response for all providers.
type Result struct {
	Completion string
	Model      string
}

// StreamResult holds the streaming response.
type StreamResult struct {
	CompletionStream io.ReadCloser
	Error            <-chan error
}

// ChatOptions holds the common options for all providers.
type ChatOptions struct {
	Temperature float32
	TopP        float32
}

// Provider is an interface that all LLM providers must implement.
type Provider interface {
	Name() string // e.g., "openai", "ollama"
	ListModels(ctx context.Context) ([]string, error)
	Chat(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (Result, error)
	ChatStream(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (StreamResult, error)
}