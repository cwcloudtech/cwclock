package telemetry

import (
	"context"
	"log/slog"
	"os"
	"time"

	otellog "go.opentelemetry.io/otel/log"
)

// stdHandler always writes structured logs to stdout, except Error+ records
// which go to stderr instead, satisfying "logs must be written in
// stdout/stderr" regardless of whether OTEL export is configured.
type stdHandler struct {
	stdout slog.Handler
	stderr slog.Handler
}

func newStdHandler(level slog.Level) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	return &stdHandler{
		stdout: slog.NewJSONHandler(os.Stdout, opts),
		stderr: slog.NewJSONHandler(os.Stderr, opts),
	}
}

func (h *stdHandler) target(level slog.Level) slog.Handler {
	if level >= slog.LevelError {
		return h.stderr
	}
	return h.stdout
}

func (h *stdHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.stdout.Enabled(ctx, level)
}

func (h *stdHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.target(r.Level).Handle(ctx, r)
}

func (h *stdHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &stdHandler{stdout: h.stdout.WithAttrs(attrs), stderr: h.stderr.WithAttrs(attrs)}
}

func (h *stdHandler) WithGroup(name string) slog.Handler {
	return &stdHandler{stdout: h.stdout.WithGroup(name), stderr: h.stderr.WithGroup(name)}
}

// multiHandler fans every log record out to each of its handlers, so app
// code logs once through the standard slog API and it reaches stdout/stderr
// and (when configured) OTEL at the same time.
type multiHandler struct {
	handlers []slog.Handler
}

func newMultiHandler(handlers []slog.Handler) slog.Handler {
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		next[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: next}
}

// otelSlogHandler bridges slog records into the OpenTelemetry Logs API so
// they're exported through the same OTLP pipeline as traces.
type otelSlogHandler struct {
	logger otellog.Logger
	attrs  []slog.Attr
}

func (h *otelSlogHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *otelSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	var rec otellog.Record
	rec.SetTimestamp(r.Time)
	rec.SetBody(otellog.StringValue(r.Message))
	rec.SetSeverity(severityFor(r.Level))
	rec.SetSeverityText(r.Level.String())

	for _, a := range h.attrs {
		rec.AddAttributes(otellog.KeyValue{Key: a.Key, Value: otelValue(a.Value)})
	}
	r.Attrs(func(a slog.Attr) bool {
		rec.AddAttributes(otellog.KeyValue{Key: a.Key, Value: otelValue(a.Value)})
		return true
	})

	h.logger.Emit(ctx, rec)
	return nil
}

func (h *otelSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Attr, 0, len(h.attrs)+len(attrs))
	next = append(next, h.attrs...)
	next = append(next, attrs...)
	return &otelSlogHandler{logger: h.logger, attrs: next}
}

func (h *otelSlogHandler) WithGroup(_ string) slog.Handler { return h }

func severityFor(level slog.Level) otellog.Severity {
	switch {
	case level >= slog.LevelError:
		return otellog.SeverityError1
	case level >= slog.LevelWarn:
		return otellog.SeverityWarn1
	case level >= slog.LevelInfo:
		return otellog.SeverityInfo1
	default:
		return otellog.SeverityDebug1
	}
}

func otelValue(v slog.Value) otellog.Value {
	switch v.Kind() {
	case slog.KindBool:
		return otellog.BoolValue(v.Bool())
	case slog.KindInt64:
		return otellog.Int64Value(v.Int64())
	case slog.KindFloat64:
		return otellog.Float64Value(v.Float64())
	case slog.KindDuration:
		return otellog.StringValue(v.Duration().String())
	case slog.KindTime:
		return otellog.StringValue(v.Time().Format(time.RFC3339Nano))
	default:
		return otellog.StringValue(v.String())
	}
}
