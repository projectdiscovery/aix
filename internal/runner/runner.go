package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/projectdiscovery/gologger"
)

var ApiUrl = "https://api.openai.com/v1/chat/completions"

type ChatGptRequest struct {
	Model    string           `json:"model"`
	Messages []ChatGptMessage `json:"messages"`
}

type ChatGptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatGptResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

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
func (r *Runner) Run() error {
	var model string
	if r.options.OpenaiApiKey == "" {
		gologger.Info().Msgf("OPENAI_API_KEY is not configured / provided.")
		os.Exit(1)
	}

	if r.options.Gpt3 {
		model = "gpt-3.5-turbo"
	}
	if r.options.Gpt4 {
		gologger.Warning().Msg("not implemented")
		os.Exit(1)
	}

	messages := []ChatGptMessage{
		{
			Role:    "user",
			Content: r.options.Prompt,
		},
	}
	reqData := ChatGptRequest{
		Model:    model,
		Messages: messages,
	}

	chatGptResp, err := r.DoGptApiRequest(reqData)
	if err != nil {
		return err
	}

	if len(chatGptResp.Choices) == 0 {
		return fmt.Errorf("no data on response")
	}

	if r.options.Verbose {
		gologger.Verbose().Msgf("[prompt] %s", messages[0].Content)
		gologger.Verbose().Msgf("[completion] %s", chatGptResp.Choices[0].Message.Content)
		return nil
	}

	if r.options.Jsonl {
		result := Result{
			Timestamp:  time.Now().String(),
			Model:      model,
			Prompt:     messages[0].Content,
			Completion: chatGptResp.Choices[0].Message.Content,
		}
		gologger.Silent().Msgf("%s", result.JSON())
		return nil
	}
	if r.options.Silent {
		gologger.Silent().Msgf("%s", chatGptResp.Choices[0].Message.Content)
		return nil
	}
	gologger.Info().Msgf("%s", chatGptResp.Choices[0].Message.Content)

	return nil
}

func (r *Runner) DoGptApiRequest(requestBody ChatGptRequest) (ChatGptResponse, error) {
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return ChatGptResponse{}, err
	}

	// Create an HTTP client and set up the API call
	client := &http.Client{}
	req, err := http.NewRequest("POST", ApiUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		return ChatGptResponse{}, err
	}

	// Add required headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.options.OpenaiApiKey))

	// Make the API call
	resp, err := client.Do(req)
	if err != nil {
		return ChatGptResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ChatGptResponse{}, fmt.Errorf("status code was %v", resp.StatusCode)
	}

	// Read and parse the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatGptResponse{}, err
	}
	var chatGptResp ChatGptResponse
	err = json.Unmarshal(respBody, &chatGptResp)
	if err != nil {
		return ChatGptResponse{}, err
	}

	return chatGptResp, nil
}
