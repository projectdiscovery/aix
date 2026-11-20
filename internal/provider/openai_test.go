package provider

import (
	"encoding/json"
	"testing"
)

func TestNewOpenAIDefaultModel(t *testing.T) {
	p, err := NewOpenAI("test-key", "")
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}
	if p.model != "gpt-6-luna" {
		t.Fatalf("default model = %q, want %q", p.model, "gpt-6-luna")
	}
}

// Reasoning models reject sampling and max_tokens params, so the default request must not carry them.
func TestCreateChatCompletionRequestOmitsUnsetParams(t *testing.T) {
	p, err := NewOpenAI("test-key", "")
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}
	body, err := json.Marshal(p.createChatCompletionRequest("hi", "", false, ChatOptions{}))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, key := range []string{"temperature", "top_p", "max_tokens", "logprobs", "top_logprobs"} {
		if _, ok := fields[key]; ok {
			t.Errorf("request body has %q: %s", key, body)
		}
	}
}
