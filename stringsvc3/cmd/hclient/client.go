package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Simple client using net/http std package
func main() {
	fUpperCase := flag.Bool("uppercase", false, "Call uppercase operation")
	fCount := flag.Bool("count", false, "Call count operation")
	flag.Parse()
	if flag.NFlag() != 1 {
		flag.Usage()
		return
	}
	reqText := strings.Join(flag.Args(), " ")
	if reqText == "" {
		reqText = "Sample Text"
	}

	tHost := "http://localhost:8080/"
	c := http.Client{Timeout: time.Duration(4) * time.Second}

	var err error
	if *fUpperCase {
		var respText string
		respText, err = callUpperCase(c, tHost, reqText)
		if err == nil {
			fmt.Printf("Succcess: uppercase result %v\n", respText)
		}

	} else if *fCount {
		var respCount int
		respCount, err = callCount(c, tHost, reqText)
		if err == nil {
			fmt.Printf("Succcess: count result %v\n", respCount)
		}
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}

func callCount(c http.Client, tHost string, reqText string) (int, error) {
	reqBody := countRequest{S: reqText}
	log.Printf("Calling uppercase with: %v", reqBody.S)

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize request with error %w", err)
	}

	reqBodyReader := bytes.NewBuffer(reqBodyJson)
	tResource := tHost + "count"
	resp, err := c.Post(tResource, "application/json", reqBodyReader)
	if err != nil {
		return 0, fmt.Errorf("failed to perform HTTP POST with error %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("received unexpected status code %v", resp.StatusCode)
	}

	respBodyJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body error %w", err)
	}

	var respBody countResponse
	err = json.Unmarshal(respBodyJson, &respBody)
	if err != nil {
		return 0, fmt.Errorf("failed to unserialize response body error %w", err)
	}

	return respBody.V, nil
}

func callUpperCase(c http.Client, tHost string, reqText string) (string, error) {
	reqBody := uppercaseRequest{S: reqText}
	log.Printf("Calling uppercase with: %v", reqBody.S)

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to serialize request with error %w", err)
	}

	reqBodyReader := bytes.NewBuffer(reqBodyJson)
	tResource := tHost + "uppercase"
	resp, err := c.Post(tResource, "application/json", reqBodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP POST with error %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received unexpected status code %v", resp.StatusCode)
	}

	respBodyJson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body error %w", err)
	}

	var respBody uppercaseResponse
	err = json.Unmarshal(respBodyJson, &respBody)
	if err != nil {
		return "", fmt.Errorf("failed to unserialize response body error %w", err)
	}

	if respBody.Err != "" {
		return "", fmt.Errorf("received error from server %v", respBody.Err)
	}

	return respBody.V, nil
}

type uppercaseRequest struct {
	S string `json:"s"`
}

type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"`
}

type countRequest struct {
	S string `json:"s"`
}

type countResponse struct {
	V int `json:"v"`
}
