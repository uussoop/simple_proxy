package errors

import "net/http"

var (
	InvalidApiKeyError = OpenAIError{
		Message:    "Incorrect API key provided",
		Type:       "invalid_request_error",
		Code:       "invalid_api_key",
		StatusCode: http.StatusUnauthorized,
		Param:      nil,
	}
)
