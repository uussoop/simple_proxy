package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

func Deflate_gzip(r *http.Response) []byte {

	var reader io.ReadCloser
	var err error
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(r.Body)

	default:
		reader = r.Body
	}
	defer reader.Close()
	resp_body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("error reading body: %s\n", err)
	}
	return resp_body
}
