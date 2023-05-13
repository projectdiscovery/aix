package runner

import (
	"encoding/json"
	"io"

	"github.com/projectdiscovery/gologger"
)

type Result struct {
	Timestamp        string         `json:"timestamp"`
	Prompt           string         `json:"prompt"`
	Completion       string         `json:"completion"`
	Model            string         `json:"model"`
	CompletionStream *io.PipeReader `json:"-"` // contained stream response
	streamWriter     *io.PipeWriter `json:"-"` // only used for streaming
	Error            error          `json:"-"`
}

// SetupStreaming sets up the streaming for the result
func (r *Result) SetupStreaming() {
	r.CompletionStream, r.streamWriter = io.Pipe()
}

// WriteCompletionStreamResponse writes a response to the completion stream
func (r *Result) WriteCompletionStreamResponse(data string) {
	r.Completion += data
	if r.streamWriter != nil && r.CompletionStream != nil {
		_, _ = r.streamWriter.Write([]byte(data))
	}
}

// CloseCompletionStream closes the completion stream
func (r *Result) CloseCompletionStream() {
	if r.streamWriter != nil && r.CompletionStream != nil {
		_ = r.streamWriter.Close()
	}
}

func (result *Result) JSON() string {
	data, err := json.Marshal(result)
	if err != nil {
		gologger.Error().Msgf("failed to marshal result: %v", err)
	}
	return string(data)
}
