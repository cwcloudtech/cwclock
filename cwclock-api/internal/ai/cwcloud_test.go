package ai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cwclock-api/internal/models"
)

func cwcloudConn(url string) models.AIConnection {
	return models.AIConnection{Format: models.AIFormatCWCloud, BaseURL: url, Model: "claude-sonnet", Token: "tok", MaxTokens: 500}
}

func TestCWCloudEndpointTolerance(t *testing.T) {
	cases := map[string]string{
		"https://api.cwcloud.tech/v1/ai/prompt": "https://api.cwcloud.tech/v1/ai/prompt",
		"https://api.cwcloud.tech/v1":           "https://api.cwcloud.tech/v1/ai/prompt",
		"https://api.cwcloud.tech":              "https://api.cwcloud.tech/v1/ai/prompt",
		"https://api.cwcloud.tech/":             "https://api.cwcloud.tech/v1/ai/prompt",
	}
	for in, want := range cases {
		if got := cwcloudEndpoint(in); got != want {
			t.Errorf("cwcloudEndpoint(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCWCloudHappyStringAndRequestShape(t *testing.T) {
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Auth-Token") != "tok" {
			t.Errorf("missing X-Auth-Token, got %q", r.Header.Get("X-Auth-Token"))
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusCreated) // 201
		_, _ = w.Write([]byte(`{"status":"ok","response":"Hello there","model":"claude-sonnet-5-5","usage":{"prompt":8,"completion":10,"total":18,"cached":2}}`))
	}))
	defer srv.Close()

	resp, err := New().askCWCloud(context.Background(), cwcloudConn(srv.URL), "claude-sonnet", Request{
		System:   "be brief",
		Messages: []Turn{{Role: models.AIChatRoleUser, Text: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.Message != "Hello there" {
		t.Errorf("message = %q", resp.Message)
	}
	if resp.Model != "claude-sonnet-5-5" {
		t.Errorf("model = %q (want CWCloud's actual model)", resp.Model)
	}
	if resp.Tokens.Input != 8 || resp.Tokens.Output != 10 || resp.Tokens.Cached != 2 {
		t.Errorf("tokens = %+v (want 8/10/2)", resp.Tokens)
	}
	if gotBody["adapter"] != "claude-sonnet" {
		t.Errorf("adapter = %v (want the model)", gotBody["adapter"])
	}
	if gotBody["enable_history"] != false {
		t.Errorf("enable_history = %v (want false)", gotBody["enable_history"])
	}
	msgs, _ := gotBody["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("messages = %v (want system + user)", gotBody["messages"])
	}
	first := msgs[0].(map[string]any)
	if first["role"] != "system" || first["message"] != "be brief" {
		t.Errorf("first message = %v (want leading system)", first)
	}
	if _, hasContent := first["content"]; hasContent {
		t.Errorf("turn used `content`, must be `message`")
	}
}

func TestCWCloudResponseArray(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"status":"ok","response":["line one","line two"],"usage":{"prompt":1,"completion":1}}`))
	}))
	defer srv.Close()
	resp, err := New().askCWCloud(context.Background(), cwcloudConn(srv.URL), "m", Request{Messages: []Turn{{Role: "user", Text: "x"}}})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if resp.Message != "line one\nline two" {
		t.Errorf("array join = %q", resp.Message)
	}
}

func TestCWCloudKoOn200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK) // 2xx but ko
		_, _ = w.Write([]byte(`{"status":"ko","error":"rate limited","i18n_code":"x","http_code":429}`))
	}))
	defer srv.Close()
	_, err := New().askCWCloud(context.Background(), cwcloudConn(srv.URL), "m", Request{Messages: []Turn{{Role: "user", Text: "x"}}})
	var refused ErrRefused
	if !errors.As(err, &refused) {
		t.Fatalf("want ErrRefused, got %v", err)
	}
	if !strings.Contains(refused.Detail, "rate limited") {
		t.Errorf("detail = %q", refused.Detail)
	}
}

func TestCWCloud424UpstreamUnwrap(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusFailedDependency) // 424
		_, _ = w.Write([]byte(`{"status":"ko","error":"model failed","upstream":{"http_code":401,"source":"anthropic","adapter":"claude-sonnet","body":"Incorrect API key"}}`))
	}))
	defer srv.Close()
	_, err := New().askCWCloud(context.Background(), cwcloudConn(srv.URL), "m", Request{Messages: []Turn{{Role: "user", Text: "x"}}})
	var refused ErrRefused
	if !errors.As(err, &refused) {
		t.Fatalf("want ErrRefused, got %v", err)
	}
	if !strings.Contains(refused.Detail, "Incorrect API key") || !strings.Contains(refused.Detail, "401") {
		t.Errorf("424 detail did not unwrap upstream: %q", refused.Detail)
	}
}

func TestCWCloudSupportsToolsPredicate(t *testing.T) {
	if models.AIFormatSupportsTools(models.AIFormatCWCloud) {
		t.Error("cwcloud must not support tools")
	}
	if !models.AIFormatSupportsTools(models.AIFormatOpenAI) {
		t.Error("openai must support tools")
	}
}
