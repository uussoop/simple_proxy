package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/uussoop/simple_proxy/config"
	"github.com/uussoop/simple_proxy/utils"
)

var api_key string = utils.Getenv("OPENAI_API_KEY", "")

var domain string = utils.Getenv("OPENAI_DOMAIN", "api.openai.com")

func Forwarder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fmt.Printf("context: %s\n", ctx.Value("config").(*config.Config).APIKeys)
	body, error := io.ReadAll(r.Body)

	fmt.Printf(string(body) + "\n")
	if error != nil {
		fmt.Printf("error reading body: %s\n", error)
	}
	path := path.Clean(r.URL.Path)
	// use differ

	defer r.Body.Close()

	req, err := http.NewRequest(strings.ToUpper(r.Method), "https://"+domain+path, bytes.NewBuffer(body))
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
	client := &http.Client{Timeout: 50 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("error making request: %s\n", err)
	}

	// resp_body := utils.Deflate_gzip(resp)
	resp_body, err := io.ReadAll(resp.Body)

	fmt.Printf("response body: %s\n", (resp_body))

	for k, v := range resp.Header {

		w.Header().Add(k, v[0])
		fmt.Printf("%s: %s\n", k, v[0])

	}
	w.WriteHeader(resp.StatusCode)
	io.WriteString(w, string(resp_body))
}
