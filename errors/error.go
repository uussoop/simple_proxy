package errors

type OpenAIError struct {
	Message    string  `json:"message"`
	Param      *string `json:"param"`
	Type       string  `json:"type"`
	Code       string  `json:"code"`
	StatusCode int     `json:"-"`
}

func (e *OpenAIError) Error() string {
	return e.Message
}
