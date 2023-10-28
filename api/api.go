package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

type streamRequest struct {
	Stream bool `json:"stream"`
	// Add other fields of the request body if applicable
}

var api_key *string

var domain *string

func Forwarder(w http.ResponseWriter, r *http.Request) {
	e := r.Context().Value(utils.EndpointKey).(*database.Endpoint)
	m := r.Context().Value(utils.ModelKey).(*database.Model)

	api_key = &e.Token
	domain = &e.Url

	defer r.Body.Close()
	authenticationToken := r.Header.Get("Authorization")
	users, exists := database.Authenticate(&authenticationToken)
	l := true
	if exists {
		u := &users[0]
		l = u.IsLimited()
	}
	// fmt.Printf("%s", users)
	logrus.Debug(exists, l)
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

		req, err := http.NewRequest(strings.ToUpper(r.Method), *domain+path, bytes.NewBuffer(bodyCopy))
		for k, v := range r.Header {
			if k == "Authorization" {
				req.Header.Add(k, "Bearer "+*api_key)
			} else {
				req.Header.Add(k, v[0])
			}
		}

		logrus.Info(req)

		if err != nil {
			fmt.Print("error in creating new request: ", err)
			return
		}

		req.Header.Add("Access-Control-Allow-Origin", "*")
		client := &http.Client{Timeout: 0}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %s\n", err)
			ServerNotReady(w)
			e.DeActivate()
			return
		}

		userEndpointModel := database.EndpointModelUsage{}

		userEndpointModel.GetOrCreate(users[0], *e, *m)

		// resp_body := utils.Deflate_gzip(resp)

		if streamBody.Stream {
			StreamResponser(&bodyCopy, w, resp, &users[0], &userEndpointModel)

		} else {

			NonStreamResponser(&bodyCopy, w, resp, &users[0], &userEndpointModel)
		}
	}
}
func NormalResponse(w http.ResponseWriter, r *http.Request, exists bool) {
	bodyCopy, readErr := io.ReadAll(r.Body)
	if readErr != nil {
		fmt.Printf("error reading body: %s\n", readErr)
	}
	path := path.Clean(r.URL.Path)
	req, err := http.NewRequest(strings.ToUpper(r.Method), *domain+path, bytes.NewBuffer(bodyCopy))

	logrus.Info(req)

	if err != nil {
		fmt.Print("error in creating new request: ", err)
	}

	for k, v := range r.Header {
		if k == "Authorization" {
			if exists {
				req.Header.Add(k, "Bearer "+*api_key)
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
	NormalStreamResponser(resp, w)
}
