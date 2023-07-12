package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/utils"
)

func updateUsageRequest(body *[]byte, user *database.User) {

	isgzip := false
	isreq := true
	requestString, requestStringerror := unmarshalOpenaiContent(body, isgzip, isreq)
	if requestStringerror != nil {
		fmt.Println(requestStringerror)
		fmt.Printf("error decoding1 :  %s\n", &requestStringerror)
	}
	var requestStringCount int
	for _, value := range requestString {

		newRequestStringCount, newReqscErr := utils.Count_tokens(value)
		requestStringCount += newRequestStringCount
		if newReqscErr != nil {
			fmt.Println(newReqscErr)
		}
	}
	fmt.Printf("request count %s \n", strconv.Itoa(requestStringCount))
	*&user.UsageToday = *&user.UsageToday + requestStringCount
	database.UpdateUserUsageToday(*user)

}
func updateUsage(resp *http.Response, resp_body *[]byte, user *database.User) {
	isgzip := false

	if resp.Header.Get("Content-Encoding") == "gzip" {
		isgzip = true
	} else {
		isgzip = false
	}
	responseString, responseStringerror := unmarshalOpenaiContent(resp_body, isgzip, false)
	if responseStringerror != nil {
		fmt.Println(responseStringerror)
		fmt.Printf("error decoding2 :  %s\n", &responseStringerror)
	}
	if responseString != nil {
		var responseStringCount int
		for _, value := range responseString {
			newResponseStringCount, newResscErr := utils.Count_tokens(value)
			responseStringCount += newResponseStringCount
			if newResscErr != nil {
				fmt.Println(newResscErr)
			}

		}
		fmt.Printf("response count %s \n", strconv.Itoa(responseStringCount))

		*&user.UsageToday = *&user.UsageToday + responseStringCount

		database.UpdateUserUsageToday(*user)
	}
}

func unmarshalOpenaiContent(body *[]byte, gzip bool, req bool) ([]string, error) {

	var responseBody interface{}
	var deflatedBody []byte
	if gzip {
		deflatedBody = utils.Deflate_gzip_byte(*body)
	} else {
		deflatedBody = *body
	}
	var splited [][]byte
	if deflatedBody == nil {
		return nil, errors.New("body is nil")
	}
	if strings.Contains("[DONE]", string(deflatedBody)) {
		return nil, errors.New("end of stream")
	}
	if strings.Contains("data:", string(deflatedBody)) {

		for _, split := range strings.Split(string(deflatedBody), "data:") {
			// Convert the string to []byte
			bytes := []byte(split)

			// Append the []byte to the splited slice
			fmt.Printf("bytes %s", split)
			splited = append(splited, bytes)
		}
	} else {
		splited = append(splited, deflatedBody)
	}
	fmt.Printf("splited %s \n", splited)
	var contents []string
	for _, bytes := range splited {
		streamErr := json.Unmarshal(bytes, &responseBody)
		if streamErr != nil {
			fmt.Println(streamErr)

		}
		// check if path choices[0].messages.content or path choices[0].delta.content
		// if not return error
		// else return content
		fmt.Printf("%s", responseBody)
		if responseBody.(map[string]interface{})["error"] != nil {

			fmt.Sprintf("error openai :  %s\n", streamErr)
		}
		if req {
			choices := responseBody.(map[string]interface{})["messages"].([]interface{})
			if choices[0].(map[string]interface{})["content"] != nil {
				contents = append(contents, choices[0].(map[string]interface{})["content"].(string))

			}
		}
		if responseBody.(map[string]interface{})["choices"] != nil {
			choices := responseBody.(map[string]interface{})["choices"].([]interface{})
			if choices[0].(map[string]interface{})["delta"] != nil {
				if choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"] != nil {
					contents = append(contents, choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"].(string))

				}
				if choices[0].(map[string]interface{})["messages"] != nil {
					if choices[0].(map[string]interface{})["messages"].(map[string]interface{})["content"] != nil {
						contents = append(contents, choices[0].(map[string]interface{})["messages"].(map[string]interface{})["content"].(string))

					}
				}
			}
		}
	}
	if len(contents) == 0 {
		return contents, errors.New("unable to get contents")

	} else {
		return contents, nil

	}
}

func NonStreamResponser(body *[]byte, w http.ResponseWriter, resp *http.Response, user *database.User) {
	updateUsageRequest(body, user)
	resp_body, err := io.ReadAll(resp.Body)
	updateUsage(resp, &resp_body, user)
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
	updateUsageRequest(body, user)

	// read resp body chunk chunk until EOF and io write each chunk if available
	for { // read chunk

		buf := make([]byte, 4*1024)
		n, err := resp.Body.Read(buf)
		buffer := buf[:n]
		updateUsage(resp, &buffer, user)
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
