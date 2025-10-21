package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/projectdiscovery/aix/internal/provider"
	errorutil "github.com/projectdiscovery/utils/errors"
)

// ErrNoKey is returned when no key is provided
var ErrNoKey = errorutil.New("OPENAI_API_KEY is not configured / provided.")

// Runner contains the internal logic of the program
type Runner struct {
	options  *Options
	provider provider.Provider
}

// NewRunner instance
func NewRunner(options *Options) (*Runner, error) {
	var p provider.Provider
	var err error

	switch options.Provider {
	case "openai":
		p, err = provider.NewOpenAI(options.OpenaiApiKey, options.Model)
	default:
		err = fmt.Errorf("unsupported provider `%s`", options.Provider)
	}

	if err != nil {
		return nil, err
	}

	return &Runner{
		options:  options,
		provider: p,
	}, nil
}

// Run the instance
func (r *Runner) Run() (*Result, error) {
	if r.options.ListModels {
		return r.runListModels()
	}

	if r.options.Prompt == "" {
		return nil, fmt.Errorf("no prompt provided")
	}

	if r.options.Stream {
		return r.runChatStream()
	}
	return r.runChat()
}

func (r *Runner) runListModels() (*Result, error) {
	models, err := r.provider.ListModels(context.Background())
	if err != nil {
		return &Result{}, err
	}

	// Use a buffer to build the output
	var buff bytes.Buffer

	if r.provider.Name() == "openai" {
		// Categorize models into gpt and o1
		var gptModels, o1Models []string
		for _, model := range models {
			switch {
			case strings.HasPrefix(model, "gpt") || strings.HasPrefix(model, "chatgpt"):
				gptModels = append(gptModels, model)
			case strings.HasPrefix(model, "o1"):
				o1Models = append(o1Models, model)
			}
		}
		// Print GPT models
		buff.WriteString("## GPT Models:\n\n")
		printModelsInGrid(&buff, gptModels, 2) // Print in 2 columns
		buff.WriteString("\n")

		// Print o1 models
		buff.WriteString("## o1 Models:\n\n")
		printModelsInGrid(&buff, o1Models, 2) // Print in 2 columns
		buff.WriteString("\n")
	} else {
		// Print models in a grid for other providers
		buff.WriteString(fmt.Sprintf("## %s Models:\n\n", r.provider.Name()))
		printModelsInGrid(&buff, models, 2) // Print in 2 columns
		buff.WriteString("\n")
	}

	result := &Result{
		Timestamp: time.Now().String(),
		Model:     r.provider.Name(),
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

func (r *Runner) runChat() (*Result, error) {
	systemPrompt := strings.Join(r.options.System, "\n")
	opts := r.newChatOptions()

	providerResult, err := r.provider.Chat(context.TODO(), r.options.Prompt, systemPrompt, opts)
	if err != nil {
		return &Result{Error: err}, err
	}
	result := &Result{
		Timestamp:  time.Now().String(),
		Model:      providerResult.Model,
		Prompt:     r.options.Prompt,
		Completion: providerResult.Completion,
	}
	return result, nil
}

func (r *Runner) runChatStream() (*Result, error) {
	systemPrompt := strings.Join(r.options.System, "\n")
	opts := r.newChatOptions()

	result := &Result{
		Timestamp: time.Now().String(),
		Model:     r.provider.Name(),
		Prompt:    r.options.Prompt,
	}

	result.SetupStreaming()
	go func(res *Result) {
		defer res.CloseCompletionStream()
		streamResult, err := r.provider.ChatStream(context.TODO(), r.options.Prompt, systemPrompt, opts)
		if err != nil {
			res.Error = err
			return
		}
		if _, err := io.Copy(res.streamWriter, streamResult.CompletionStream); err != nil {
			res.Error = err
			return
		}
	}(result)
	return result, nil
}

func (r *Runner) newChatOptions() provider.ChatOptions {
	return provider.ChatOptions{
		Temperature: r.options.Temperature,
		TopP:        r.options.TopP,
	}
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
		fmt.Fprintf(buff, "%-*s", columnWidth, model)
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
