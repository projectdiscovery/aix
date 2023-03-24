package runner

import "encoding/json"

type Result struct {
	Timestamp  string `json:"timestamp"`
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Model      string `json:"model"`
}

func (result *Result) JSON() string {
	data, _ := json.Marshal(result)
	return string(data)
}
