package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/uussoop/simple_proxy/config"
	"github.com/uussoop/simple_proxy/utils"
)

type streamRequest struct {
	Stream bool `json:"stream"`
	// Add other fields of the request body if applicable
}

var api_key string = utils.Getenv("OPENAI_API_KEY", "")

var domain string = utils.Getenv("OPENAI_DOMAIN", "api.openai.com")

func Forwarder(w http.ResponseWriter, r *http.Request) {
	var streamBody streamRequest
	bodyCopy, readErr := io.ReadAll(r.Body)          // Create a copy of the request body
	r.Body = io.NopCloser(bytes.NewBuffer(bodyCopy)) // Restore the request body with the copy

	streamErr := json.Unmarshal(bodyCopy, &streamBody) // Decode the copied body
	if streamErr != nil {
		// Handle JSON decoding error
		fmt.Printf("Failed to decode JSON: %s\n", streamErr)
		return
	}

	ctx := r.Context()
	fmt.Printf("context: %s\n", ctx.Value("config").(*config.Config).APIKeys)
	fmt.Printf("Body: %s\n", string(bodyCopy))
	if readErr != nil {
		fmt.Printf("error reading body: %s\n", readErr)
	}
	path := path.Clean(r.URL.Path)
	// use differ

	defer r.Body.Close()

	req, err := http.NewRequest(strings.ToUpper(r.Method), "https://"+domain+path, bytes.NewBuffer(bodyCopy))
	for k, v := range r.Header {
		if k == "Authorization" {
			for client, ky := range ctx.Value("config").(*config.Config).APIKeys {
				fmt.Printf("found client: \n%s\n%s ", client, ky)

				if strings.Contains(v[0], ky) {
					fmt.Printf("found client: \n%s\n%s ", client, ky)
					req.Header.Add(k, "Bearer "+api_key)
				}
			}
			if req.Header.Get("Authorization") == "" {
				req.Header.Add(k, v[0])
			}
		} else {
			req.Header.Add(k, v[0])
			fmt.Printf("%s: %s\n", k, v[0])
		}

	}
	req.Header.Add("Access-Control-Allow-Origin", "*")
	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
	}

	// resp_body := utils.Deflate_gzip(resp)

	// Check if the value of the "stream" field is true
	if streamBody.Stream {

		// read resp body chunk chunk until EOF and io write each chunk if available
		for { // read chunk
			fmt.Printf("chunk length: %s\n", strconv.FormatInt(resp.ContentLength, 10))
			buf := make([]byte, 4*1024)
			n, err := resp.Body.Read(buf)
			fmt.Printf(string(buf[:n]))
			if n == 0 {
				break
			}
			if err != nil && err != io.EOF {
				fmt.Printf("error reading response body: %s\n", err)
				break
			}
			for k, v := range resp.Header {

				w.Header().Add(k, v[0])
				fmt.Printf("%s: %s\n", k, v[0])

			}
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Write(buf[:n])
		}
	} else {
		resp_body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s\n", err)
		}
		fmt.Printf("response body: %s\n", (resp_body))

		for k, v := range resp.Header {

			w.Header().Add(k, v[0])
			fmt.Printf("%s: %s\n", k, v[0])

		}

		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.WriteString(w, string(resp_body))
	}
}
