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

	"github.com/uussoop/simple_proxy/database"
	"github.com/uussoop/simple_proxy/pkg/cache"
	"github.com/uussoop/simple_proxy/utils"
)

func updateUsageRequest(body *[]byte, user *database.User) {
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
	// fmt.Printf("request count %s \n", strconv.Itoa(requestStringCount))
	// user.UsageToday = user.UsageToday + requestStringCount
	cach := cache.GetCache()
	nowadder, ok := cach.Get(strconv.Itoa(int(user.ID)))
	if !ok {
		cach.Set(strconv.Itoa(int(user.ID)), requestStringCount+10, 0)
	} else {

		cach.Set(strconv.Itoa(int(user.ID)), ((nowadder.(int)) + requestStringCount + 10), 0)

	}
	// database.UpdateUserUsageToday(*&user.ID, requestStringCount, false)

}
func updateUsage(resp *http.Response, resp_body *[]byte, user *database.User) {
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
		// fmt.Printf("response count %s \n", strconv.Itoa(responseStringCount))

		// user.UsageToday = user.UsageToday + responseStringCount
		cach := cache.GetCache()
		nowadder, ok := cach.Get(strconv.Itoa(int(user.ID)))
		if !ok {
			cach.Set(strconv.Itoa(int(user.ID)), responseStringCount+10, 0)
		} else {

			cach.Set(strconv.Itoa(int(user.ID)), ((nowadder.(int)) + responseStringCount + 10), 0)

		}
		// database.UpdateUserUsageToday(*&user.ID, responseStringCount, false)
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
		// fmt.Printf("hiiiiiiiiiiiiiiii %s \n", string(bytes))

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

			messages, ok := responseBody.(map[string]interface{})["messages"].([]interface{})
			if ok {
				for _, message := range messages {

					if message.(map[string]interface{})["content"] != nil {
						mes, ok := message.(map[string]interface{})["content"].(string)
						if ok {
							contents = append(
								contents,
								mes,
							)
							continue
						}
						c, ok := message.(map[string]interface{})["content"].([]interface{})

						if ok {

							for _, content := range c {

								if tp, ok := content.(map[string]interface{})["type"]; ok {
									if tp.(string) == "text" {
										text, ok := content.(map[string]interface{})["text"].(string)
										if ok {

											contents = append(contents, text)
										}
									}
									if tp.(string) == "image_url" {

										contents = append(contents, utils.ImageTokenString)

									}
								}

							}
						}

					}
				}
			}
			emb, ok := responseBody.(map[string]interface{})["input"].(string)
			if ok {
				contents = append(
					contents,
					emb,
				)
			}

		}
		if responseBody.(map[string]interface{})["choices"] != nil {
			choices := responseBody.(map[string]interface{})["choices"].([]interface{})

			if choices[0].(map[string]interface{})["delta"] != nil {
				if choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"] != nil {
					contents = append(
						contents,
						choices[0].(map[string]interface{})["delta"].(map[string]interface{})["content"].(string),
					)

				}

			}
			if choices[0].(map[string]interface{})["message"] != nil {

				if choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"] != nil {
					contents = append(
						contents,
						choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string),
					)

				}
			}
			if choices[0].(map[string]interface{})["text"] != nil {

				if choices[0].(map[string]interface{})["text"].(map[string]interface{})["content"] != nil {
					contents = append(
						contents,
						choices[0].(map[string]interface{})["text"].(string),
					)

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

func NonStreamResponser(
	body *[]byte,
	w http.ResponseWriter,
	resp *http.Response,
	user *database.User,

) {

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
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(resp.StatusCode)

	io.WriteString(w, string(resp_body))
}

func StreamResponser(
	body *[]byte,
	w http.ResponseWriter,
	resp *http.Response,
	user *database.User,
) {

	go updateUsageRequest(body, user)
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')

		if err == io.EOF {
			break
		}

		go updateUsage(resp, &line, user)
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
