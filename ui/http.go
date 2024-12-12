package ui

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"

	"net/http"
	"time"
)

var Verbose = false

func POST(url string, authToken string, headers map[string]string, content interface{}) ([]byte, error) {
	return RPC("POST", url, authToken, headers, content)
}

func GET(url string, authToken string, headers map[string]string) ([]byte, error) {
	return RPC("GET", url, authToken, headers, nil)
}

func DELETE(url string, authToken string, headers map[string]string) ([]byte, error) {
	return RPC("DELETE", url, authToken, headers, nil)
}

func RPC(method string, url string, authToken string, headers map[string]string, content interface{}) ([]byte, error) {
	var postBody io.Reader = nil
	var bodyBytes []byte
	if content != nil {
		bodyBytes, _ = json.Marshal(content)
		postBody = bytes.NewBuffer(bodyBytes)
	}

	req, err := http.NewRequest(method, url, postBody)
	if err != nil {
		log.Println("Error crating http request. ", err)
		return nil, err
	}
	if Verbose {
		log.Printf("%s %s\n", method, url)
	}

	req.Header.Set("Cache-Control", "no-cache")
	if content != nil {
		req.Header.Set("Content-Type", "application/json")
		//log.Printf("%s: %s\n", "Content-Type", "application/json")
	}

	if authToken != "" {
		req.Header.Set("Authorization", authToken)
		//log.Printf("%s: %s\n", "Authorization", authToken)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
		//log.Printf("%s: %s\n", k, v)
	}

	if Verbose && bodyBytes != nil {
		log.Printf("%s\n", string(bodyBytes))
	}

	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error reading response. ", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Fatal("Error reading body. ", err)
		return nil, err
	}

	if Verbose {
		log.Printf("%s\n", string(body))
	}
	return body, nil
}
