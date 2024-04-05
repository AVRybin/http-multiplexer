package httpClientMultiplexer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type ResponseSingleUrl struct {
	URL  string      `json:"url"`
	Body interface{} `json:"body"`
}

func SendMultiplexerRequests(ctx context.Context, urlList []string, method string,
	timeout time.Duration, maxParallelReq int) ([]ResponseSingleUrl, error) {

	size := len(urlList)
	result := make([]ResponseSingleUrl, 0, size)

	_, cancel := context.WithCancel(ctx)
	sem := make(chan struct{}, maxParallelReq)

	errors := make(chan error, 1)
	results := make(chan ResponseSingleUrl, size)

	for _, url := range urlList {
		sem <- struct{}{}

		go func(url string) {
			body, err := sendHTTPRequest(ctx, method, url, []byte{}, timeout)
			if err != nil {
				errors <- err
				cancel()
			}

			results <- body
			<-sem
		}(url)
	}

	for i := 0; i < size; i++ {
		select {
		case r := <-results:
			result = append(result, r)
		case err := <-errors:
			cancel()
			return nil, err
		}
	}

	close(errors)
	close(results)

	cancel()
	return result, nil
}

func sendHTTPRequest(ctx context.Context, method string, url string, body []byte,
	timeout time.Duration) (ResponseSingleUrl, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))

	if err != nil {
		return ResponseSingleUrl{}, err
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return ResponseSingleUrl{}, err
	}

	if resp.StatusCode != 200 {
		return ResponseSingleUrl{}, errors.New("not 200 code in response")
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseSingleUrl{}, err
	}

	var bodyInterface interface{}
	err = json.Unmarshal(responseBody, &bodyInterface)

	if err != nil {
		return ResponseSingleUrl{}, err
	}

	resp.Body.Close()
	return ResponseSingleUrl{URL: url, Body: bodyInterface}, nil
}
