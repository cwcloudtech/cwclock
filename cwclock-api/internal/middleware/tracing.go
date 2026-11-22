package middleware

import (
	"context"
	"cwclock-api/internal/utils"
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

// ErrorFunc is an optional hook fired once for every 5xx response, used to raise a
// Discord [ERROR] alert. It carries the request's correlation id and the server
// error's detail (the cause the client is NOT shown), so the alert is
// actionable. It runs after the response is written, so it must post
// asynchronously and never block. May be nil (disabled).
type ErrorFunc func(method, route string, status int, cid, detail string)

// ServerErrorReporter is implemented by the response recorder: a handler stashes
// the raw detail of a 5xx on it (writeError), so the [ERROR] alert can carry it
// while the client only ever sees a generic message.
type ServerErrorReporter interface {
	ReportServerError(detail string)
}

// Instrument starts a tracing span for every request, logs one line per
// request via logger, and (if observe is non-nil) reports the call to
// observe for metrics. The span/log/metrics all use the same resolved chi
// route pattern: opened with the raw path (not resolved yet at that point),
// then renamed once the inner handler chain returns and
// chi.RouteContext has settled on the matched pattern (e.g.
// "/organizations/{orgId}/clients/{clientId}") instead of the raw path, so
// they group by endpoint rather than by ID.
func Instrument(tracer trace.Tracer, logger *slog.Logger, observe EndpointObserver, onError ErrorFunc) func(http.Handler) http.Handler {
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
				if p := rc.RoutePattern(); utils.IsNotBlank(p) {
					route = p
				}
			}

			span.SetName(r.Method + " " + route)
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.route", route),
				attribute.Int("http.status_code", rec.status),
			)
			if cid := CIDFromContext(r.Context()); utils.IsNotBlank(cid) {
				span.SetAttributes(attribute.String("cid", cid))
			}
			if rec.status >= 500 {
				span.SetStatus(codes.Error, http.StatusText(rec.status))
			}

			logRequest(ctx, logger, r.Method, route, rec.status, duration)

			if observe != nil {
				observe(route, r.Method, rec.status, duration)
			}
			if onError != nil && rec.status >= 500 {
				onError(r.Method, route, rec.status, CIDFromContext(r.Context()), rec.serverError)
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
	attrs := []any{
		slog.String("method", method),
		slog.String("route", route),
		slog.Int("status", status),
		slog.Duration("duration", duration),
	}
	if cid := CIDFromContext(ctx); utils.IsNotBlank(cid) {
		attrs = append(attrs, slog.String("cid", cid))
	}
	logger.Log(ctx, level, "http request", attrs...)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
	// serverError is the detail of a 5xx the handler chose not to disclose to the
	// client (set via ReportServerError), surfaced only in the [ERROR] alert.
	serverError string
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// ReportServerError records the undisclosed detail of a 5xx for the alert hook.
func (r *statusRecorder) ReportServerError(detail string) {
	r.serverError = detail
}
