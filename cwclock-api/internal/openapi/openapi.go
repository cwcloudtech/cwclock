// Package openapi builds an OpenAPI 3.0 document dynamically from the live
// chi router (chi.Walk over its registered routes) instead of a hand-
// maintained spec file, so it can never drift out of sync with the API.
package openapi

import (
	"cwclock-api/internal/utils"
	"net/http"
	"reflect"
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
	Summary     string                `json:"summary,omitempty"`
	Tags        []string              `json:"tags,omitempty"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	RequestBody *RequestBody          `json:"requestBody,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Security    []map[string][]string `json:"security,omitempty"`
}

type Parameter struct {
	Name     string  `json:"name"`
	In       string  `json:"in"`
	Required bool    `json:"required"`
	Schema   *Schema `json:"schema"`
}

// EndpointDoc is the hand-maintained request shape of a route — the parts the
// router walk cannot derive: the JSON body struct, and the query / header
// parameters the handler reads. The body schema is still reflected from the
// struct (see schemaOf), so only the route→doc mapping
// (handlers.EndpointDocs) is maintained by hand.
type EndpointDoc struct {
	Body   reflect.Type
	Query  []Param
	Header []Param
}

// Param is one query or header parameter. Type defaults to "string".
type Param struct {
	Name     string
	Required bool
	Type     string
}

// RequestBody is the application/json body of a write endpoint, its schema
// reflected from the handler's payload struct.
type RequestBody struct {
	Required bool                 `json:"required,omitempty"`
	Content  map[string]MediaType `json:"content"`
}

type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

// Schema is a (subset of) JSON Schema: a type, an optional format, object
// properties and array items — enough to describe the request-body structs and
// the simple path/query/header parameters.
type Schema struct {
	Type       string             `json:"type,omitempty"`
	Format     string             `json:"format,omitempty"`
	Properties map[string]*Schema `json:"properties,omitempty"`
	Items      *Schema            `json:"items,omitempty"`
}

type Response struct {
	Description string `json:"description"`
}

type Components struct {
	SecuritySchemes map[string]SecurityScheme `json:"securitySchemes,omitempty"`
}

type SecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme,omitempty"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	In           string `json:"in,omitempty"`
	Name         string `json:"name,omitempty"`
}

var pathParam = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// publicRoutes don't require a bearer token, unlike the rest of the API.
var publicRoutes = map[string]bool{
	"GET /v1/health":       true,
	"GET /v1/manifest":     true,
	"GET /v1/currencies":   true,
	"POST /v1/users/":      true,
	"POST /v1/users/login": true,
	"GET /metrics":         true,
}

// Generate walks every route currently registered on r and describes it as
// an OpenAPI operation: method, path, path parameters, a resource tag taken
// from the first segment after /v1, and an auth requirement (bearer JWT or
// X-Api-Key, either satisfies it - mirroring middleware.Auth) for every
// route except the known-public ones (login/register/currencies).
//
// docs maps "METHOD /route" to the request shape the walk can't derive: the
// JSON body struct (whose schema is reflected, so it tracks the struct), and
// the query / header parameters the handler reads. Only that route→doc mapping
// (see handlers.EndpointDocs) is maintained by hand.
func Generate(r chi.Router, title, version string, docs map[string]EndpointDoc) Spec {
	paths := map[string]PathItem{}

	_ = chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		if route == "/" || route == "/openapi.json" {
			return nil
		}

		// chi emits "/v1/foo/" for a sub-router's index route ("/"), which
		// shows up inconsistently next to the slash-less routes. Drop the
		// trailing slash for the path emitted in the spec; the docs and
		// public-route lookups below still key on the original chi route.
		specPath := route
		if trimmed := strings.TrimRight(specPath, "/"); trimmed != "" {
			specPath = trimmed
		}

		item, ok := paths[specPath]
		if !ok {
			item = PathItem{}
			paths[specPath] = item
		}

		var params []Parameter
		for _, m := range pathParam.FindAllStringSubmatch(route, -1) {
			params = append(params, Parameter{Name: m[1], In: "path", Required: true, Schema: &Schema{Type: "string"}})
		}

		op := Operation{
			Summary:    method + " " + specPath,
			Tags:       []string{tagFor(route)},
			Parameters: params,
			Responses:  map[string]Response{"200": {Description: "OK"}},
		}
		if doc, ok := docs[method+" "+route]; ok {
			for _, q := range doc.Query {
				op.Parameters = append(op.Parameters, Parameter{
					Name: q.Name, In: "query", Required: q.Required, Schema: &Schema{Type: paramType(q.Type)},
				})
			}
			for _, hp := range doc.Header {
				op.Parameters = append(op.Parameters, Parameter{
					Name: hp.Name, In: "header", Required: hp.Required, Schema: &Schema{Type: paramType(hp.Type)},
				})
			}
			if doc.Body != nil {
				op.RequestBody = &RequestBody{
					Required: true,
					Content:  map[string]MediaType{"application/json": {Schema: schemaOf(doc.Body)}},
				}
			}
		}
		if !publicRoutes[method+" "+route] {
			op.Security = []map[string][]string{{"bearerAuth": {}}, {"apiKeyAuth": {}}}
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
				"apiKeyAuth": {Type: "apiKey", In: "header", Name: "X-Api-Key"},
			},
		},
	}
}

// tagFor groups routes for the Swagger UI sidebar by the first path segment
// after the /v1 prefix (e.g. "/v1/organizations/{orgId}/reports" -> "organizations").
func tagFor(route string) string {
	segments := strings.Split(utils.GetUriPath(route), "/")
	if len(segments) >= 2 && segments[0] == "v1" {
		return segments[1]
	}
	if len(segments) >= 1 && utils.IsNotBlank(segments[0]) {
		return segments[0]
	}
	return "default"
}

// schemaOf derives a JSON Schema from a Go type (a request payload struct),
// following the same rules encoding/json does: field names come from the
// `json` tag, `json:"-"` fields and unexported fields are dropped, and an
// anonymous (embedded) struct's fields are inlined. Nested structs, slices,
// maps and pointers recurse; a self-referential struct is cut off as a bare
// object.
func schemaOf(t reflect.Type) *Schema { return schemaOfSeen(t, map[reflect.Type]bool{}) }

func schemaOfSeen(t reflect.Type, seen map[reflect.Type]bool) *Schema {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	switch t.String() {
	case "time.Time":
		return &Schema{Type: "string", Format: "date-time"}
	case "json.RawMessage":
		return &Schema{} // arbitrary JSON
	}
	switch t.Kind() {
	case reflect.String:
		return &Schema{Type: "string"}
	case reflect.Bool:
		return &Schema{Type: "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &Schema{Type: "integer"}
	case reflect.Float32, reflect.Float64:
		return &Schema{Type: "number"}
	case reflect.Slice, reflect.Array:
		if t.Elem().Kind() == reflect.Uint8 { // []byte → base64 string in JSON
			return &Schema{Type: "string"}
		}
		return &Schema{Type: "array", Items: schemaOfSeen(t.Elem(), seen)}
	case reflect.Map:
		return &Schema{Type: "object"}
	case reflect.Interface:
		return &Schema{} // any
	case reflect.Struct:
		if seen[t] {
			return &Schema{Type: "object"}
		}
		seen[t] = true
		defer delete(seen, t)
		s := &Schema{Type: "object", Properties: map[string]*Schema{}}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name, skip, inline := jsonFieldName(f)
			if skip || (f.PkgPath != "" && !f.Anonymous) {
				continue
			}
			if inline {
				ft := f.Type
				for ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					for k, v := range schemaOfSeen(ft, seen).Properties {
						s.Properties[k] = v
					}
				}
				continue
			}
			s.Properties[name] = schemaOfSeen(f.Type, seen)
		}
		return s
	}
	return &Schema{}
}

// jsonFieldName reads a struct field's json tag: the property name (or the Go
// field name when untagged), whether it is skipped (`json:"-"`), and whether it
// is an untagged embedded struct whose fields should be inlined.
func jsonFieldName(f reflect.StructField) (name string, skip, inline bool) {
	tag := f.Tag.Get("json")
	first := tag
	if i := strings.IndexByte(tag, ','); i >= 0 {
		first = tag[:i]
	}
	if first == "-" {
		return "", true, false
	}
	if first == "" {
		if f.Anonymous {
			return "", false, true
		}
		return f.Name, false, false
	}
	return first, false, false
}

// paramType defaults a query/header parameter's schema type to "string".
func paramType(t string) string {
	if t == "" {
		return "string"
	}
	return t
}
