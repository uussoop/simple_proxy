package errors

import "net/http"

var (
	MaxQuotaError = OpenAIError{
		Message:    "You exceeded your current quota, please check your plan and billing details.",
		Type:       "insufficient_quota",
		Code:       "insufficient_quota",
		Param:      nil,
		StatusCode: http.StatusTooManyRequests,
	}

	RateLimitError = OpenAIError{
		Message:    "Rate limit reached for default-gpt-3.5-turbo in organization org-kxTIJdD5jFZpdZ2X9YvBy4hZ on requests per min. Limit: 3 / min. Please try again in 20s. Contact us through our help center at help.openai.com if you continue to have issues. Please add a payment method to your account to increase your rate limit. Visit https://platform.openai.com/account/billing to add a payment method.",
		Type:       "requests",
		Param:      nil,
		Code:       "rate_limit_reached",
		StatusCode: http.StatusTooManyRequests,
	}
)
