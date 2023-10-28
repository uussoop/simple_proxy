package errors

import (
	"net/http"
)

var (
	BadRequestError = OpenAIError{
		Message:    "Bad Request",
		Type:       "invalid_request_error",
		Code:       "bad_request",
		StatusCode: http.StatusBadRequest,
	}
)
