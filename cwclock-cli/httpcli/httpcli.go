package httpcli

import (
	"bytes"
	"cwclock/config"
	"cwclock/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
}

func (e ErrorResponse) message() string {
	if utils.IsNotBlank(e.Message) {
		return e.Message
	}
	return e.Error
}

func RequestPath(path string) string {
	hostname := strings.TrimRight(config.GetApiURL(), "/")
	if path == "/metrics" {
		return hostname + path
	}
	return fmt.Sprintf("%s/v1%s", hostname, path)
}

func HttpRequest(cli *http.Client, path string, method string, body bytes.Buffer, headers ...map[string]string) (closer io.ReadCloser, err error) {
	req, err := http.NewRequest(method, RequestPath(path), &body)
	if nil != err {
		return nil, err
	}

	if apiKey := config.GetApiKey(); utils.IsNotBlank(apiKey) {
		req.Header.Set("X-Api-Key", apiKey)
	}

	requestHeaders := make(map[string]string)
	if len(headers) > 0 && headers[0] != nil {
		for key, value := range headers[0] {
			requestHeaders[key] = value
		}
	}

	if len(requestHeaders) == 0 {
		requestHeaders["Content-Type"] = "application/json"
	}

	for key, value := range requestHeaders {
		req.Header.Set(key, value)
	}

	resp, err := cli.Do(req)
	if nil != err {
		return nil, err
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		return resp.Body, nil
	case resp.StatusCode >= 400:
		resp_body := new(bytes.Buffer)
		_, err := resp_body.ReadFrom(resp.Body)
		if nil != err {
			return nil, errors.New("an error occurred")
		}

		rawBody := strings.TrimSpace(resp_body.String())
		errorResponse := ErrorResponse{}
		json.NewDecoder(resp_body).Decode(&errorResponse)
		if utils.IsBlank(errorResponse.message()) {
			if utils.IsNotBlank(rawBody) {
				return nil, fmt.Errorf("server error (status %d): %s", resp.StatusCode, rawBody)
			}
			return nil, fmt.Errorf("server error with status %d", resp.StatusCode)
		} else {
			return nil, errors.New(errorResponse.message())
		}
	}

	return nil, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
}
