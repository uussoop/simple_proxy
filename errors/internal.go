package errors

import "net/http"

var (
	InternalError = OpenAIError{
		Message:    "Internal server error",
		Type:       "internal_error",
		Code:       "internal_error",
		StatusCode: http.StatusInternalServerError,
		Param:      nil,
	}

	OverloadedError = OpenAIError{
		Message:    "Server is overloaded please try again later",
		Type:       "internal_error",
		Code:       "internal_error",
		StatusCode: http.StatusServiceUnavailable,
		Param:      nil,
	}

	ServerNotReadyError = OpenAIError{
		Message:    "Server not ready!",
		Type:       "internal_error",
		Code:       "internal_error",
		StatusCode: http.StatusServiceUnavailable,
		Param:      nil,
	}
)
