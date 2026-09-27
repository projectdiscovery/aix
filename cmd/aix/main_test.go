package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type streamTransport func(*http.Request) (*http.Response, error)

func (f streamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestStreamingCLI(t *testing.T) {
	if mode := os.Getenv("AIX_TEST_STREAM_MODE"); mode != "" {
		http.DefaultTransport = streamTransport(func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost || req.URL.String() != "http://localhost:11434/api/generate" {
				t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
			}
			body := "{\"response\":\"partial response\",\"done\":false}\n"
			switch mode {
			case "success":
				body += "{\"response\":\" finished\",\"done\":true}\n"
			case "generation error":
				body += "{\"error\":\"model generation failed\"}\n"
			case "invalid response":
				body += "invalid json\n"
			case "request failure":
				return nil, errors.New("cannot reach model server")
			default:
				t.Fatalf("unknown test mode %q", mode)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/x-ndjson"}},
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    req,
			}, nil
		})
		// Explicit options avoid update checks, terminal probing and loading
		// a user's saved provider credentials. All HTTP is intercepted above.
		os.Args = []string{"aix", "-provider", "ollama", "-m", "test", "-s", "-duc", "-nm", "-nc", "-ak", "unused-test-key"}
		main()
		return
	}

	for _, tc := range []struct {
		mode     string
		wantText string
		wantCode int
	}{
		{mode: "success", wantText: "partial response finished"},
		{mode: "generation error", wantText: "model generation failed", wantCode: 1},
		{mode: "invalid response", wantText: "failed to decode streaming response", wantCode: 1},
		{mode: "request failure", wantText: "cannot reach model server", wantCode: 1},
	} {
		t.Run(tc.mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStreamingCLI$")
			cmd.Env = append(os.Environ(), "AIX_TEST_STREAM_MODE="+tc.mode)
			cmd.Stdin = strings.NewReader("test prompt")
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("streaming command did not exit: %s", output)
			}
			code := 0
			if err != nil {
				var exitErr *exec.ExitError
				if !errors.As(err, &exitErr) {
					t.Fatal(err)
				}
				code = exitErr.ExitCode()
			}
			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d; output: %s", code, tc.wantCode, output)
			}
			if !strings.Contains(string(output), tc.wantText) {
				t.Errorf("output missing %q: %s", tc.wantText, output)
			}
		})
	}
}
