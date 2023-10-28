package api

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/rodrikv/openai_proxy/database"
	"github.com/rodrikv/openai_proxy/utils"
	"github.com/sirupsen/logrus"
)

func updateUsageRequest(body *[]byte, user *database.User, userEndpointUsage *database.EndpointModelUsage) (token int) {
	isgzip := false
	isreq := true
	requestString, requestStringerror := unmarshalOpenaiContent(body, isgzip, isreq)
	if requestStringerror != nil {
		fmt.Println(requestStringerror)

	}
	var requestStringCount int
	for _, value := range requestString {

		newRequestStringCount, newReqscErr := utils.Count_tokens(value)
		requestStringCount += newRequestStringCount
		if newReqscErr != nil {
			fmt.Println(newReqscErr)
		}
	}
	logrus.Debugf("request count %s \n", strconv.Itoa(requestStringCount))
	user.UsageToday = user.UsageToday + requestStringCount

	database.UpdateUserUsageToday(*user)
	userEndpointUsage.Increase(uint(requestStringCount))

	token = requestStringCount
	return
}
func updateUsage(resp *http.Response, resp_body *[]byte, user *database.User, userEndpointUsage *database.EndpointModelUsage) (token int) {
	isgzip := false

	if resp.Header.Get("Content-Encoding") == "gzip" {
		isgzip = true
	} else {
		isgzip = false
	}
	if !strings.Contains(string(*resp_body), "data:") {
		return
	}
	responseString, responseStringerror := unmarshalOpenaiContent(resp_body, isgzip, false)
	if responseStringerror != nil {
		fmt.Println(responseStringerror)
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

		user.UsageToday = user.UsageToday + responseStringCount

		database.UpdateUserUsageToday(*user)
		userEndpointUsage.Increase(uint(responseStringCount))
		token = responseStringCount
	}
	return
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
	if len(string(deflatedBody)) == 0 {
		return nil, errors.New("body is empty")
	}
	if strings.Contains(string(deflatedBody), "[DONE]") {
		return nil, errors.New("end of stream")
	}

	if strings.Contains(string(deflatedBody), "data:") && !req {

		for _, split := range strings.Split(string(deflatedBody), "data:") {
			// Convert the string to []byte
			bytes := []byte(split)

			// Append the []byte to the splited slice

			if len(bytes) != 0 {

				splited = append(splited, bytes)
			}
		}
	} else {

		splited = append(splited, deflatedBody)
	}
	var contents []string
	for _, bytes := range splited {
		fmt.Printf("hiiiiiiiiiiiiiiii %s \n", string(bytes))

		streamErr := json.Unmarshal(bytes, &responseBody)
		if streamErr != nil {
			fmt.Println(streamErr)

		}

		// check if path choices[0].messages.content or path choices[0].delta.content
		// if not return error
		// else return content
		if responseBody.(map[string]interface{})["error"] != nil {

			fmt.Println(streamErr)
		}
		if req {
			prompt, pok := responseBody.(map[string]interface{})["prompt"].(string)
			message, ok := responseBody.(map[string]interface{})["messages"].([]interface{})
			if ok {
				if message[0].(map[string]interface{})["content"] != nil {
					contents = append(contents, message[0].(map[string]interface{})["content"].(string))

				}
			} else if pok {

				contents = append(contents, prompt)

			}

		}
		if responseBody.(map[string]interface{})["choices"] != nil {
			choices := responseBody.(map[string]interface{})["choices"].([]interface{})

			if choices[0].(map[string]interface{})["delta"] != nil {
				if choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"] != nil {
					contents = append(contents, choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"].(string))

				}

			}
			if choices[0].(map[string]interface{})["message"] != nil {

				if choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"] != nil {
					contents = append(contents, choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string))

				}
			}
			if choices[0].(map[string]interface{})["text"] != nil {

				if choices[0].(map[string]interface{})["text"].(map[string]interface{})["content"] != nil {
					contents = append(contents, choices[0].(map[string]interface{})["text"].(string))

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

func NonStreamResponser(body *[]byte, w http.ResponseWriter, resp *http.Response, user *database.User, userEndpointUsage *database.EndpointModelUsage) {
	t := updateUsageRequest(body, user, userEndpointUsage)

	if t > 4000 {
		MaxContextError(w)
		return
	}

	resp_body, err := io.ReadAll(resp.Body)
	updateUsage(resp, &resp_body, user, userEndpointUsage)
	// database.UpdateUser(database.User{Token: authenticationToken, UsageToday: users[0].UsageToday + requestStringCount + responseStringCount})
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
	}

	for k, v := range resp.Header {

		w.Header().Add(k, v[0])

	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)

	io.WriteString(w, string(resp_body))
}

func StreamResponser(body *[]byte, w http.ResponseWriter, resp *http.Response, user *database.User, userEndpointUsage *database.EndpointModelUsage) {
	t := updateUsageRequest(body, user, userEndpointUsage)

	if t > 4000 {
		MaxContextError(w)
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		go updateUsage(resp, &line, user, userEndpointUsage)
		for k, v := range resp.Header {

			w.Header().Add(k, v[0])
		}
		if err != nil {
			fmt.Printf("error reading response body stream: %s\n", err)

		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, writeErr := w.Write(line)
		if writeErr != nil {
			fmt.Printf("error writing response body: %s\n", writeErr)
			break
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func NormalStreamResponser(resp *http.Response, w http.ResponseWriter) {
	resp_body, err := io.ReadAll(resp.Body)

	// database.UpdateUser(database.User{Token: authenticationToken, UsageToday: users[0].UsageToday + requestStringCount + responseStringCount})
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
	}

	for k, v := range resp.Header {

		w.Header().Add(k, v[0])

	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)

	io.WriteString(w, string(resp_body))
}
