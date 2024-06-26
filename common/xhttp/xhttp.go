package xhttp

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// 发送 HTTP 请求
func sendRequest(method, url string, data []byte) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "token 5b59c8a43dfb287afba7df5ae145f41e9b26a537")
	req.Header.Set("Content-Type", "application/json")
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Post POST 请求
func Post(url string, data []byte) (string, error) {
	return sendRequest("POST", url, data)
}

// Put PUT 请求
func Put(url string, data []byte) (string, error) {
	return sendRequest("PUT", url, data)
}

// Delete DELETE 请求
func Delete(url string, data []byte) (string, error) {
	return sendRequest("DELETE", url, data)
}

// Head HEAD 请求
func Head(url string, data []byte) (string, error) {
	return sendRequest("HEAD", url, data)
}

// Connect CONNECT 请求
func Connect(url string, data []byte) (string, error) {
	return sendRequest("CONNECT", url, data)
}

// Options OPTIONS 请求
func Options(url string, data []byte) (string, error) {
	return sendRequest("OPTIONS", url, data)
}

// TRACE 请求
func Trace(url string, data []byte) (string, error) {
	return sendRequest("TRACE", url, data)
}
