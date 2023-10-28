package api

import (
	"encoding/json"
	"net/http"

	"github.com/rodrikv/openai_proxy/errors"
)

func MaxQuotaError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.MaxQuotaError.StatusCode)

	rateLimitErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.MaxQuotaError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(rateLimitErrorJson)
}

func RateLimitError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.RateLimitError.StatusCode)

	rateLimitErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.RateLimitError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(rateLimitErrorJson)
}

func MaxContextError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.MaxContextError.StatusCode)

	maxContextErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.MaxContextError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(maxContextErrorJson)
}

func InvalidApiKeyError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.InvalidApiKeyError.StatusCode)

	invalidApiKeyErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.InvalidApiKeyError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(invalidApiKeyErrorJson)
}

func ServerIsOverloaded(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.OverloadedError.StatusCode)

	serverIsOverloadedJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.OverloadedError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(serverIsOverloadedJson)
}

func ServerNotReady(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.ServerNotReadyError.StatusCode)

	serverIsOverloadedJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.ServerNotReadyError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(serverIsOverloadedJson)
}

func InternalServerError(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.InternalError.StatusCode)

	internalServerErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.InternalError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(internalServerErrorJson)
}

func BadRequest(w http.ResponseWriter, text string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.BadRequestError.StatusCode)

	badRequestErrorJson, err := json.Marshal(map[string]errors.OpenAIError{"error": errors.BadRequestError})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(badRequestErrorJson)
}
