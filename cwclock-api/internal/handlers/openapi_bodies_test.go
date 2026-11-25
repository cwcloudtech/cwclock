package handlers

import (
	"net/http"
	"strings"
	"testing"

	"cwclock-api/internal/openapi"

	"github.com/go-chi/chi/v5"
)

// specForDocs generates the spec the way the router does — chi.Walk over real
// route patterns, with the real EndpointDocs() map — so these assertions are
// about what /openapi.json actually serves, not about a fixture.
func specForDocs(t *testing.T) openapi.Spec {
	t.Helper()
	noop := func(w http.ResponseWriter, r *http.Request) {}
	r := chi.NewRouter()
	r.Route("/v1/ai", func(r chi.Router) {
		r.Post("/chat", noop)            // documents a body, no answer
		r.Post("/cwcloud/adapter", noop) // documents the whole exchange
	})
	r.Get("/v1/health", noop) // documents nothing at all
	return openapi.Generate(r, "CWClock API", "0.0.0-test", EndpointDocs())
}

func operation(t *testing.T, spec openapi.Spec, path, method string) openapi.Operation {
	t.Helper()
	item, ok := spec.Paths[path]
	if !ok {
		t.Fatalf("%s is missing from the spec", path)
	}
	op, ok := item[method]
	if !ok {
		t.Fatalf("%s has no %s operation", path, method)
	}
	return op
}

// jsonSchema is the application/json schema of a response, or a failure saying
// the response documents no body.
func jsonSchema(t *testing.T, op openapi.Operation, code string) *openapi.Schema {
	t.Helper()
	response, ok := op.Responses[code]
	if !ok {
		t.Fatalf("no %q response documented", code)
	}
	media, ok := response.Content["application/json"]
	if !ok {
		t.Fatalf("the %q response documents no application/json body", code)
	}
	if media.Schema == nil {
		t.Fatalf("the %q response has an application/json entry with no schema", code)
	}
	return media.Schema
}

// The CWAI adapter is written to somebody else's contract, so its reply has to
// be in the spec: whoever registers the adapter's URL in CWCloud reads
// /openapi.json and never this repository.
func TestAdapterDocumentsItsAnswer(t *testing.T) {
	op := operation(t, specForDocs(t), "/v1/ai/cwcloud/adapter", "post")

	// The request body was always documented; it stays.
	if op.RequestBody == nil {
		t.Fatal("the adapter documents no request body")
	}

	answer := jsonSchema(t, op, "200")
	if answer.Type != "object" {
		t.Fatalf("the answer schema is %q, want an object", answer.Type)
	}
	// The fields the CWAI contract makes a caller's parser read, reflected from
	// cwcloudAdapterResponse — so the spec tracks the struct the handler marshals.
	for _, field := range []string{"status", "message", "usage"} {
		if answer.Properties[field] == nil {
			t.Errorf("the answer schema has no %q property", field)
		}
	}
	usage := answer.Properties["usage"]
	if usage == nil {
		t.Fatal("the answer schema has no usage property")
	}
	for _, field := range []string{"prompt_tokens", "completion_tokens", "total_tokens"} {
		count := usage.Properties[field]
		if count == nil {
			t.Errorf("usage has no %q property", field)
			continue
		}
		if count.Type != "integer" {
			t.Errorf("usage.%s is %q, want an integer", field, count.Type)
		}
	}

	// A description of its own: "OK" is what told the integrator nothing.
	description := op.Responses["200"].Description
	if description == "OK" || strings.TrimSpace(description) == "" {
		t.Errorf("the answer still carries the generic description %q", description)
	}
	if !strings.Contains(description, "message") || !strings.Contains(description, "usage") {
		t.Errorf("the answer's description says nothing about message/usage: %q", description)
	}
}

// A refusal carries cwclock's ordinary error body, under "default" because one
// shape covers every status the handler can refuse with.
func TestAdapterDocumentsItsRefusal(t *testing.T) {
	op := operation(t, specForDocs(t), "/v1/ai/cwcloud/adapter", "post")

	failure := jsonSchema(t, op, "default")
	if failure.Type != "object" {
		t.Fatalf("the refusal schema is %q, want an object", failure.Type)
	}
	for _, field := range []string{"message", "i18n_code"} {
		if failure.Properties[field] == nil {
			t.Errorf("the refusal schema has no %q property", field)
		}
	}
	if failure.Properties["status"] != nil || failure.Properties["usage"] != nil {
		t.Error("the refusal is documented as the CWAI answer shape; cwclock refuses with its own error body")
	}

	description := op.Responses["default"].Description
	if strings.TrimSpace(description) == "" || description == "Error" {
		t.Errorf("the refusal carries no description of its own: %q", description)
	}
	// The status is part of what a caller has to read, so the sentence says so.
	for _, code := range []string{"400", "401"} {
		if !strings.Contains(description, code) {
			t.Errorf("the refusal's description does not mention the %s status: %q", code, description)
		}
	}
}

// Every other route keeps the bare 200. A generator that started inventing
// response schemas would be worse than one documenting nothing, so this is the
// assertion that matters most: the adapter is the exception, not the new rule.
func TestRoutesWithoutAnAnswerKeepTheBare200(t *testing.T) {
	spec := specForDocs(t)
	for _, route := range []struct{ path, method string }{
		{"/v1/ai/chat", "post"}, // documents a request body only
		{"/v1/health", "get"},   // documents nothing
	} {
		op := operation(t, spec, route.path, route.method)
		if len(op.Responses) != 1 {
			t.Errorf("%s %s documents %d responses, want only the bare 200", route.method, route.path, len(op.Responses))
		}
		response, ok := op.Responses["200"]
		if !ok {
			t.Fatalf("%s %s has no 200 response", route.method, route.path)
		}
		if response.Description != "OK" {
			t.Errorf("%s %s describes its 200 as %q, want %q", route.method, route.path, response.Description, "OK")
		}
		if response.Content != nil {
			t.Errorf("%s %s documents a 200 body it never declared: %#v", route.method, route.path, response.Content)
		}
		if _, ok := op.Responses["default"]; ok {
			t.Errorf("%s %s documents a refusal it never declared", route.method, route.path)
		}
	}
}

// The declaration itself: exactly one route in the whole map documents a reply.
// A second one appearing here is a decision to make, not a detail to slip in.
func TestOnlyTheAdapterDocumentsAReply(t *testing.T) {
	for route, doc := range EndpointDocs() {
		if doc.Answer == nil && doc.Failure == nil {
			continue
		}
		if route != "POST /v1/ai/cwcloud/adapter" {
			t.Errorf("%s documents a reply; only the CWAI adapter contract should", route)
		}
		if doc.Answer == nil || doc.Failure == nil {
			t.Errorf("%s documents half an exchange: answer=%v failure=%v", route, doc.Answer, doc.Failure)
		}
	}
}
