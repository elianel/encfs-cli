package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	HttpClient = http.Client{
		Timeout: 10 * time.Second,
	}
	PartSize    = 7000000
	MaxRetries  = 5
	DefaultHost = "encfs.just-h.party"
)

func requestWithRetries(client *http.Client, req *http.Request, maxRetries int) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 1; i <= maxRetries; i++ {
		log.Printf("%s %s, Try %d\n", req.Method, req.URL, i)
		resp, err = client.Do(req)
		if err != nil {
			continue
		}
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}
		if resp.StatusCode != http.StatusOK {
			err = errors.New(fmt.Sprintf("%s Request %s Failure: %s\n", req.Method, req.URL, resp.Status))
		}
	}
	return nil, err
}
