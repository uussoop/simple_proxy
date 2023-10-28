package errors

import "net/http"

var (
	maxContextParam = "messages"
)

var (
	MaxContextError = OpenAIError{
		Message:    "This model's maximum context length is 4097 tokens. However, your messages resulted in 14192 tokens. Please reduce the length of the messages.",
		Type:       "invalid_request_error",
		Code:       "context_length_exceeded",
		Param:      &maxContextParam,
		StatusCode: http.StatusBadRequest,
	}
)
