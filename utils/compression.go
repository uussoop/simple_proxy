package utils

import (
	"bytes"
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
	if err != nil {
		fmt.Printf("error reading body: %s\n", err)
	}

	resp_body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("error reading body: %s\n", err)
	}
	return resp_body
}
func Deflate_gzip_byte(r []byte) []byte {

	var reader io.ReadCloser
	var err error
	reader, err = gzip.NewReader(bytes.NewReader(r))
	if err != nil {
		fmt.Printf("error reading body: %s\n", err)
	}
	defer reader.Close()
	resp_body, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("error reading body: %s\n", err)
	}
	return resp_body
}
