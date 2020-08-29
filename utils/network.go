package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: time.Duration(10 * time.Second)}

func doRequest(method, uri string, body []byte, header map[string]string) ([]byte, error) {

	req, err := http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bt, err := ioutil.ReadAll(resp.Body)

	return bt, err
}

//HTTPGet get
func HTTPGet(url string, header map[string]string) ([]byte, error) {
	return doRequest("GET", url, nil, header)
}

//HTTPGetMap map
func HTTPGetMap(url string, header map[string]string) (bodyMap map[string]interface{}, err error) {
	body, err := doRequest("GET", url, nil, header)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &bodyMap)
	return bodyMap, err
}

//HTTPPost post
func HTTPPost(url string, body []byte, header map[string]string) ([]byte, error) {
	return doRequest("POST", url, body, header)
}

//HTTPPostMap map
func HTTPPostMap(url string, post []byte, header map[string]string) (bodyMap map[string]interface{}, err error) {
	body, err := doRequest("POST", url, post, header)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &bodyMap)
	return bodyMap, err
}
