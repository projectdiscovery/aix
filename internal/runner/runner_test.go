package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/projectdiscovery/aix/internal/provider"
)

type streamingProvider struct {
	provider.Provider
	stream provider.StreamResult
	err    error
}

func (p *streamingProvider) Name() string { return "test" }

func (p *streamingProvider) ChatStream(context.Context, string, string, provider.ChatOptions) (provider.StreamResult, error) {
	return p.stream, p.err
}

func TestRunChatStreamReportsErrors(t *testing.T) {
	generationErr := errors.New("model generation failed")
	for _, tc := range []struct {
		name       string
		completion string
		startErr   error
		streamErr  error
		noErrors   bool
	}{
		{name: "success", completion: "complete response"},
		{name: "no error channel", completion: "complete response", noErrors: true},
		{name: "startup failure", startErr: generationErr},
		{name: "generation failure", streamErr: generationErr},
		{name: "partial response", completion: "partial response", streamErr: generationErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var errorChannel chan error
			if !tc.noErrors {
				errorChannel = make(chan error, 1)
				if tc.streamErr != nil {
					errorChannel <- tc.streamErr
				}
				close(errorChannel)
			}
			r := &Runner{
				options: &Options{Prompt: "test", Stream: true},
				provider: &streamingProvider{
					stream: provider.StreamResult{
						CompletionStream: io.NopCloser(strings.NewReader(tc.completion)),
						Error:            errorChannel,
					},
					err: tc.startErr,
				},
			}
			result, err := r.Run()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = result.CompletionStream.Close() })
			type completion struct {
				text string
				err  error
			}
			done := make(chan completion, 1)
			go func() {
				data, err := io.ReadAll(result.CompletionStream)
				done <- completion{text: string(data), err: err}
			}()
			select {
			case got := <-done:
				if got.text != tc.completion {
					t.Errorf("completion = %q, want %q", got.text, tc.completion)
				}
				wantErr := tc.startErr
				if wantErr == nil {
					wantErr = tc.streamErr
				}
				if !errors.Is(got.err, wantErr) {
					t.Errorf("stream error = %v, want %v", got.err, wantErr)
				}
				if !errors.Is(result.Error, wantErr) {
					t.Errorf("result error = %v, want %v", result.Error, wantErr)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("completion stream did not close")
			}
		})
	}
}
