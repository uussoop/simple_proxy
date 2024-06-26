package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/utils"
)

type streamRequest struct {
	Stream bool   `json:"stream"`
	Model  string `json:"model"`
	// Add other fields of the request body if applicable
}

var api_key string = utils.Getenv("OPENAI_API_KEY", "")

var domain string = utils.Getenv("OPENAI_DOMAIN", "api.openai.com")

func Forwarder(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	authenticationToken := r.Header.Get("Authorization")
	users, exists := database.Authenticate(&authenticationToken)
	islimited := true
	if exists {
		islimited = database.IsLimited(&users[0])
	}
	// fmt.Printf("%s", users)
	fmt.Println(exists, islimited)
	if exists && !islimited {
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

		req, err := http.NewRequest(
			strings.ToUpper(r.Method),
			"https://"+domain+path,
			bytes.NewBuffer(bodyCopy),
		)
		for k, v := range r.Header {
			if k == "Authorization" {
				req.Header.Add(k, "Bearer "+api_key)
			} else {
				req.Header.Add(k, v[0])
			}

		}
		req.Header.Add("Access-Control-Allow-Origin", "*")
		client := &http.Client{Timeout: 0}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %s\n", err)
		}

		// resp_body := utils.Deflate_gzip(resp)
		isvision := false
		isvision = strings.Contains(streamBody.Model, "vision")
		if streamBody.Stream {
			StreamResponser(&bodyCopy, w, resp, &users[0])

		} else {

			NonStreamResponser(&bodyCopy, w, resp, &users[0], isvision)
		}
	} else {
		if r.Method == "POST" {
			io.WriteString(w, `{
				"error": {
				  "message": "Quota exceeded for the requested resource. this is not openai this is beastbrain",
				  "type": "insufficient_quota",
				  "param": null,
				  "code": "quota_exceeded"
				}
			  }`)
			return
		} else {
			NormalResponse(w, r, exists)
		}
	}
}
func NormalResponse(w http.ResponseWriter, r *http.Request, exists bool) {
	bodyCopy, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		fmt.Printf("error reading body: %s\n", readErr)
	}
	path := path.Clean(r.URL.Path)
	req, err := http.NewRequest(
		strings.ToUpper(r.Method),
		"https://"+domain+path,
		bytes.NewBuffer(bodyCopy),
	)
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
	client := &http.Client{Timeout: 0}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
	}
	fmt.Println("different kind of request")
	if resp == nil {
		w.Write([]byte(err.Error()))
		return
	}
	NormalStreamResponser(resp, w)
}
