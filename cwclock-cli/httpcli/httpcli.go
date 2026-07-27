package httpcli

import (
	"bytes"
	"cwclock/config"
	"cwclock/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
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
		defer resp.Body.Close()
		return nil, readErrorResponse(resp)
	}

	return nil, fmt.Errorf("unhandled status code: %d", resp.StatusCode)
}

func readErrorResponse(resp *http.Response) error {
	resp_body := new(bytes.Buffer)
	_, err := resp_body.ReadFrom(resp.Body)
	if nil != err {
		return errors.New("an error occurred")
	}

	rawBody := strings.TrimSpace(resp_body.String())
	errorResponse := ErrorResponse{}
	json.Unmarshal(resp_body.Bytes(), &errorResponse)
	if utils.IsBlank(errorResponse.message()) {
		if utils.IsNotBlank(rawBody) {
			return fmt.Errorf("server error (status %d): %s", resp.StatusCode, rawBody)
		}
		return fmt.Errorf("server error with status %d", resp.StatusCode)
	}
	return errors.New(errorResponse.message())
}

// HttpRequestFile is HttpRequest's counterpart for binary downloads (invoice
// PDFs): it reads the full response body into memory instead of returning a
// stream, and surfaces the server-suggested filename from the
// Content-Disposition header so callers don't have to invent one.
func HttpRequestFile(cli *http.Client, path string, method string, body bytes.Buffer) (data []byte, filename string, err error) {
	req, err := http.NewRequest(method, RequestPath(path), &body)
	if nil != err {
		return nil, utils.EMPTY, err
	}

	if apiKey := config.GetApiKey(); utils.IsNotBlank(apiKey) {
		req.Header.Set("X-Api-Key", apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := cli.Do(req)
	if nil != err {
		return nil, utils.EMPTY, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, utils.EMPTY, readErrorResponse(resp)
	}

	respBody := new(bytes.Buffer)
	if _, err := respBody.ReadFrom(resp.Body); nil != err {
		return nil, utils.EMPTY, err
	}

	return respBody.Bytes(), filenameFromContentDisposition(resp.Header.Get("Content-Disposition")), nil
}

func filenameFromContentDisposition(header string) string {
	if utils.IsBlank(header) {
		return utils.EMPTY
	}
	_, params, err := mime.ParseMediaType(header)
	if nil != err {
		return utils.EMPTY
	}
	return params["filename"]
}
