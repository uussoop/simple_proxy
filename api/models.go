package api

import (
	"net/http"

	"github.com/uussoop/simple_proxy/database"
)

// Chat completion request
type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	// other fields like temperature, etc
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Chat completion response
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Image generation request
type ImageGenerationRequest struct {
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
	// other fields
}

// Image generation response
type ImageGenerationResponse struct {
	Created int64   `json:"created"`
	Data    []Image `json:"data"`
}

type Image struct {
	URL string `json:"url"`
}

// Fine tuning job request
type FineTuningJobRequest struct {
	TrainingFileID string `json:"training_file"`
	Model          string `json:"model"`
	// other fields like hyperparameters
}

// Fine tuning job response
type FineTuningJobResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Model     string `json:"model"`
	Status    string `json:"status"`
}

func updateUsageTest(resp *http.Response, body *[]byte, user *database.User, isRequest bool, endpoint string) {

	// var deflatedBody []byte

	// if resp.Header.Get("Content-Encoding") == "gzip" {

	// 	deflatedBody = utils.Deflate_gzip_byte(*body)

	// } else {
	// 	deflatedBody = *body

	// }
	// if
	// json.Unmarshal(deflatedBody)

}
