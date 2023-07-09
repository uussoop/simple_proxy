package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/utils"
)

func updateUsage(body *[]byte, resp *http.Response, resp_body *[]byte, user *database.User) {
	isgzip := false
	isreq := true
	requestString, requestStringerror := unmarshalOpenaiContent(body, isgzip, isreq)
	if requestStringerror != nil {
		fmt.Printf("error decoding1 :  %s\n", &requestStringerror)
	}
	requestStringCount, reqscErr := utils.Count_tokens(requestString)
	if reqscErr != nil {
		fmt.Printf("error decoding3 :  %s\n", &reqscErr)
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		isgzip = true
	} else {
		isgzip = false
	}
	responseString, responseStringerror := unmarshalOpenaiContent(resp_body, isgzip, false)
	if responseStringerror != nil {
		fmt.Printf("error decoding2 :  %s\n", &responseStringerror)
	}
	responseStringCount, resscErr := utils.Count_tokens(responseString)
	if resscErr != nil {
		fmt.Printf("error decoding4 :  %s\n", &resscErr)
	}
	*&user.UsageToday = *&user.UsageToday + requestStringCount + responseStringCount
	database.UpdateUserUsageToday(*user)
}

func unmarshalOpenaiContent(body *[]byte, gzip bool, req bool) (string, error) {
	fmt.Printf("body: %s\n", string(*body))
	var responseBody interface{}
	var deflatedBody []byte
	if gzip {
		deflatedBody = utils.Deflate_gzip_byte(*body)
	} else {
		deflatedBody = *body
	}
	streamErr := json.Unmarshal(deflatedBody, &responseBody)
	if streamErr != nil {
		fmt.Printf("error decoding unmarshal :  %s\n", &streamErr)
		return "", streamErr
	}
	// check if path choices[0].messages.content or path choices[0].delta.content
	// if not return error
	// else return content
	if responseBody.(map[string]interface{})["error"] != nil {
		return "", errors.New(fmt.Sprintf("error openai :  %s\n", streamErr))
	}
	if req {
		choices := responseBody.(map[string]interface{})["messages"].([]interface{})
		if choices[0].(map[string]interface{})["content"] != nil {
			return choices[0].(map[string]interface{})["content"].(string), nil

		}
	}
	if responseBody.(map[string]interface{})["choices"] != nil {
		choices := responseBody.(map[string]interface{})["choices"].([]interface{})
		if choices[0].(map[string]interface{})["delta"] != nil {
			if choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"] != nil {
				return choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"].(string), nil
			}
			if choices[0].(map[string]interface{})["messages"] != nil {
				if choices[0].(map[string]interface{})["messages"].(map[string]interface{})["content"] != nil {
					return choices[0].(map[string]interface{})["messages"].(map[string]interface{})["content"].(string), nil
				}
			}
		}
	}
	return "", errors.New(fmt.Sprintf("error decoding :  %s\n", &streamErr))
}

func NonStreamResponser(body *[]byte, w http.ResponseWriter, resp *http.Response, user *database.User) {

	resp_body, err := io.ReadAll(resp.Body)
	updateUsage(body, resp, &resp_body, user)
	// database.UpdateUser(database.User{Token: authenticationToken, UsageToday: users[0].UsageToday + requestStringCount + responseStringCount})
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
	}

	for k, v := range resp.Header {

		w.Header().Add(k, v[0])

	}

	w.WriteHeader(resp.StatusCode)

	io.WriteString(w, string(resp_body))
}
func StreamResponser(body *[]byte, w http.ResponseWriter, resp *http.Response, user *database.User) {
	// read resp body chunk chunk until EOF and io write each chunk if available
	for { // read chunk

		buf := make([]byte, 4*1024)
		n, err := resp.Body.Read(buf)
		bufToUpdateUsage := buf[:n]
		updateUsage(&bufToUpdateUsage, resp, &buf, user)
		if n == 0 {
			break
		}
		if err != nil && err != io.EOF {
			fmt.Printf("error reading response body: %s\n", err)
			break
		}
		for k, v := range resp.Header {

			w.Header().Add(k, v[0])
		}
		_, writeErr := w.Write(buf[:n])
		if writeErr != nil {
			fmt.Printf("error writing response body: %s\n", writeErr)
			break
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}
