package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/utils"
)

type streamRequest struct {
	Stream bool `json:"stream"`
	// Add other fields of the request body if applicable
}

var api_key string = utils.Getenv("OPENAI_API_KEY", "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")

var domain string = utils.Getenv("OPENAI_DOMAIN", "api.openai.com")

func Forwarder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	authenticationToken := r.Header.Get("Authorization")
	users, exists := database.Authenticate(&authenticationToken)
	l := true
	if exists {
		l = database.IsLimited(&users[0])
	}
	// fmt.Printf("%s", users)
	fmt.Println(exists, l)
	if exists && !l {
		var streamBody streamRequest
		bodyCopy, readErr := io.ReadAll(r.Body) // Create a copy of the request body
		// r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy)) // Restore the request body with the copy

		streamErr := json.Unmarshal(bodyCopy, &streamBody) // Decode the copied body
		if streamErr != nil {
			// Handle JSON decoding error
			fmt.Printf("Failed to decode JSON: %s\n", streamErr)
			NormalResponse(w, r, exists)
			return

		}

		// ctx := r.Context()

		if readErr != nil {
			fmt.Printf("error reading body: %s\n", readErr)
		}
		path := path.Clean(r.URL.Path)
		// use differ

		req, err := http.NewRequest(strings.ToUpper(r.Method), "https://"+domain+path, bytes.NewBuffer(bodyCopy))
		for k, v := range r.Header {
			if k == "Authorization" {
				req.Header.Add(k, "Bearer "+api_key)
			} else {
				req.Header.Add(k, v[0])
			}

		}
		req.Header.Add("Access-Control-Allow-Origin", "*")
		client := &http.Client{Timeout: 50 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %s\n", err)
		}

		// resp_body := utils.Deflate_gzip(resp)

		if streamBody.Stream {
			StreamResponser(&bodyCopy, w, resp, &users[0])

		} else {

			NonStreamResponser(&bodyCopy, w, resp, &users[0])
		}
	} else {
		NormalResponse(w, r, exists)
	}
}
func NormalResponse(w http.ResponseWriter, r *http.Request, exists bool) {
	bodyCopy, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		fmt.Printf("error reading body: %s\n", readErr)
	}
	path := path.Clean(r.URL.Path)
	req, err := http.NewRequest(strings.ToUpper(r.Method), "https://"+domain+path, bytes.NewBuffer(bodyCopy))
	for k, v := range r.Header {

		if k == "Authorization" {
			if exists {
				req.Header.Add(k, "Bearer "+api_key)
			} else {
				req.Header.Add(k, v[0])
			}

		} else {
			req.Header.Add(k, v[0])
		}

	}

	req.Header.Add("Access-Control-Allow-Origin", "*")
	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
	}
	fmt.Println("different kind of request")
	NormalStreamResponser(resp, w)
}
