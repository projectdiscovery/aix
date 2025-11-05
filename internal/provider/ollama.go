package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ollama/ollama/api"
)

// OllamaProvider communicates with an Ollama server.
type OllamaProvider struct {
	client  *http.Client
	baseURL string
	model   string
}

// NewOllama creates a new OllamaProvider.
func NewOllama(model string) (*OllamaProvider, error) {
	return &OllamaProvider{
		client:  &http.Client{},
		baseURL: "http://localhost:11434", // This can be made configurable later
		model:   model,
	}, nil
}

func (p *OllamaProvider) Name() string {
	return "ollama"
}

func (p *OllamaProvider) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var listResp api.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var modelIDs []string
	for _, model := range listResp.Models {
		modelIDs = append(modelIDs, model.Name)
	}
	return modelIDs, nil
}

func (p *OllamaProvider) Chat(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (Result, error) {
	reqBody := p.createGenerateRequest(prompt, systemPrompt, false, opts)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return Result{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return Result{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{}, fmt.Errorf("failed to read response body: %w", err)
	}

	var errResp struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
		return Result{}, errors.New(errResp.Error)
	}

	var genResp api.GenerateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return Result{}, fmt.Errorf("failed to decode response: %w. raw response: %s", err, string(body))
	}

	return Result{
		Completion: genResp.Response,
		Model:      genResp.Model,
	}, nil
}

func (p *OllamaProvider) ChatStream(ctx context.Context, prompt string, systemPrompt string, opts ChatOptions) (StreamResult, error) {
	reqBody := p.createGenerateRequest(prompt, systemPrompt, true, opts)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return StreamResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return StreamResult{}, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return StreamResult{}, fmt.Errorf("failed to send request: %w", err)
	}

	r, w := io.Pipe()
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if err := resp.Body.Close(); err != nil {
				errChan <- err
			}
		}()
		defer func() {
			if err := w.Close(); err != nil {
				errChan <- err
			}
		}()
		defer close(errChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var errResp struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(line, &errResp); err == nil && errResp.Error != "" {
				errChan <- errors.New(errResp.Error)
				return
			}

			var genResp api.GenerateResponse
			if err := json.Unmarshal(line, &genResp); err != nil {
				errChan <- fmt.Errorf("failed to decode streaming response: %w. raw response: %s", err, string(line))
				return
			}

			_, writeErr := w.Write([]byte(genResp.Response))
			if writeErr != nil {
				errChan <- writeErr
				return
			}
		}
		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("failed to read streaming response: %w", err)
		}
	}()

	return StreamResult{
		CompletionStream: r,
		Error:            errChan,
	}, nil
}

func (p *OllamaProvider) createGenerateRequest(prompt string, systemPrompt string, stream bool, opts ChatOptions) api.GenerateRequest {
	streamPtr := &stream
	req := api.GenerateRequest{
		Model:  p.model,
		Prompt: prompt,
		System: systemPrompt,
		Stream: streamPtr,
	}

	ollaOpts := make(map[string]any)
	if opts.Temperature != 0 {
		ollaOpts["temperature"] = opts.Temperature
	}
	if opts.TopP != 0 {
		ollaOpts["top_p"] = opts.TopP
	}
	req.Options = ollaOpts

	return req
}
