package client

import (
	"bytes"
	"cwclock/httpcli"
	"io"
	"net/http"
)

func NewClient() (*Client, error) {
	return &Client{
		httpClient: &http.Client{},
	}, nil
}

func (c *Client) httpRequest(path string, method string, body bytes.Buffer, headers ...map[string]string) (closer io.ReadCloser, err error) {
	return httpcli.HttpRequest(c.httpClient, path, method, body, headers...)
}
