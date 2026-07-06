package handlers

import (
	"net/http"

	"cwclock-api/internal/openapi"
)

// NewOpenAPIHandler serves a pre-generated OpenAPI spec as JSON. The spec is
// built once at startup (see openapi.Generate), not regenerated per request.
func NewOpenAPIHandler(spec openapi.Spec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, spec)
	}
}

const swaggerUIPage = `<!doctype html>
<html>
<head>
  <meta charset="utf-8" />
  <title>CWClock API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.json",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>
`

// ServeSwaggerUI serves a static Swagger UI page pointed at /openapi.json.
func ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerUIPage))
}
