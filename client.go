package main

import (
	"fmt"
	"net"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	pathBase   string
}

func NewClient(pathBase string) *Client {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableCompression:    false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return &Client{
		httpClient: client,
		pathBase:   pathBase,
	}
}

func (c *Client) Post(path string, payload []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented yet")
}

func (c *Client) Get(path string, payload []byte) ([]byte, error) {
	return nil, fmt.Errorf("not implemented yet")
}
