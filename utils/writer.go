package utils

import "net/http"

type StatusCaptureResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *StatusCaptureResponseWriter) WriteHeader(statusCode int) {
	w.StatusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
