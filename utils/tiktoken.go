package utils

import (
	"github.com/pkoukk/tiktoken-go"
	"github.com/sirupsen/logrus"
)

func Count_tokens(text string) (int, error) {

	encoding := "gpt-3.5-turbo"
	tke, err := tiktoken.EncodingForModel(encoding)
	if err != nil {
		return 0, err
	}

	// encode
	token := tke.Encode(text, nil, nil)

	return len(token), nil
}

func CountTokens(modelName, text string) (lenTokens int, err error) {
	var token []int
	tke, err := tiktoken.EncodingForModel(modelName)
	if err != nil {
		logrus.Error(modelName, err)
		return 0, err
	}
	token = tke.Encode(text, nil, nil)

	return len(token), nil
}
