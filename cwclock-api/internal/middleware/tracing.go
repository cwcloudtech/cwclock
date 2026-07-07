package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// EndpointObserver is notified once per completed request with the resolved
// route pattern, status and duration, so metrics can be recorded without a
// third middleware re-deriving the same information.
type EndpointObserver func(route, method string, status int, duration time.Duration)

// Instrument starts a tracing span for every request, logs one line per
// request via logger, and (if observe is non-nil) reports the call to
// observe for metrics. The span/log/metrics all use the same resolved chi
// route pattern: opened with the raw path (not resolved yet at that point),
// then renamed once the inner handler chain returns and
// chi.RouteContext has settled on the matched pattern (e.g.
// "/organizations/{orgId}/clients/{clientId}") instead of the raw path, so
// they group by endpoint rather than by ID.
func Instrument(tracer trace.Tracer, logger *slog.Logger, observe EndpointObserver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path, trace.WithSpanKind(trace.SpanKindServer))
			defer span.End()

			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))
			duration := time.Since(start)

			route := r.URL.Path
			if rc := chi.RouteContext(r.Context()); rc != nil {
				if p := rc.RoutePattern(); p != "" {
					route = p
				}
			}

			span.SetName(r.Method + " " + route)
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", route),
				attribute.Int("http.status_code", rec.status),
			)
			if rec.status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(rec.status))
			}

			logRequest(ctx, logger, r.Method, route, rec.status, duration)

			if observe != nil {
				observe(route, r.Method, rec.status, duration)
			}
		})
	}
}

func logRequest(ctx context.Context, logger *slog.Logger, method, route string, status int, duration time.Duration) {
	level := slog.LevelInfo
	if status >= 500 {
		level = slog.LevelError
	} else if status >= 400 {
		level = slog.LevelWarn
	}
	logger.Log(ctx, level, "http request",
		slog.String("method", method),
		slog.String("route", route),
		slog.Int("status", status),
		slog.Duration("duration", duration),
	)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
