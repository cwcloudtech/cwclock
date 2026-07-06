// Package openapi builds an OpenAPI 3.0 document dynamically from the live
// chi router (chi.Walk over its registered routes) instead of a hand-
// maintained spec file, so it can never drift out of sync with the API.
package openapi

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Spec struct {
	OpenAPI    string              `json:"openapi"`
	Info       Info                `json:"info"`
	Paths      map[string]PathItem `json:"paths"`
	Components Components          `json:"components"`
}

type Info struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description,omitempty"`
}

// PathItem maps an HTTP method (lowercase) to its operation.
type PathItem map[string]Operation

type Operation struct {
	Summary    string                `json:"summary,omitempty"`
	Tags       []string              `json:"tags,omitempty"`
	Parameters []Parameter           `json:"parameters,omitempty"`
	Responses  map[string]Response   `json:"responses"`
	Security   []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name     string            `json:"name"`
	In       string            `json:"in"`
	Required bool              `json:"required"`
	Schema   map[string]string `json:"schema"`
}

type Response struct {
	Description string `json:"description"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme"`
	BearerFormat string `json:"bearerFormat,omitempty"`
}

var pathParam = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// publicRoutes don't require a bearer token, unlike the rest of the API.
var publicRoutes = map[string]bool{
	"GET /v1/currencies":   true,
	"POST /v1/users/":      true,
	"POST /v1/users/login": true,
}

// Generate walks every route currently registered on r and describes it as
// an OpenAPI operation: method, path, path parameters, a resource tag taken
// from the first segment after /v1, and a bearer-auth requirement for every
// route except the known-public ones (login/register/currencies).
func Generate(r chi.Router, title, version string) Spec {
	paths := map[string]PathItem{}

	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/" || route == "/openapi.json" {
			return nil
		}

		item, ok := paths[route]
		if !ok {
			item = PathItem{}
			paths[route] = item
		}

		var params []Parameter
		for _, m := range pathParam.FindAllStringSubmatch(route, -1) {
			params = append(params, Parameter{Name: m[1], In: "path", Required: true, Schema: map[string]string{"type": "string"}})
		}

		op := Operation{
			Summary:    method + " " + route,
			Tags:       []string{tagFor(route)},
			Parameters: params,
			Responses:  map[string]Response{"200": {Description: "OK"}},
		}
		if !publicRoutes[method+" "+route] {
			op.Security = []map[string][]string{{"bearerAuth": {}}}
		}

		item[strings.ToLower(method)] = op
		return nil
	})

	return Spec{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       title,
			Version:     version,
			Description: "Dynamically generated from the live router — every registered route is reflected here automatically.",
		},
		Paths: paths,
		Components: Components{
			SecuritySchemes: map[string]SecurityScheme{
				"bearerAuth": {Type: "http", Scheme: "bearer", BearerFormat: "JWT"},
			},
		},
	}
}

// tagFor groups routes for the Swagger UI sidebar by the first path segment
// after the /v1 prefix (e.g. "/v1/organizations/{orgId}/reports" -> "organizations").
func tagFor(route string) string {
	segments := strings.Split(strings.TrimPrefix(route, "/"), "/")
	if len(segments) >= 2 && segments[0] == "v1" {
		return segments[1]
	}
	if len(segments) >= 1 && segments[0] != "" {
		return segments[0]
	}
	return "default"
}
