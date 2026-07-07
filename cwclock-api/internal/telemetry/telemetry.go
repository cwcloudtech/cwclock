// Package telemetry wires up OpenTelemetry tracing and a dual-sink logger
// (always stdout/stderr, additionally OTLP when configured) behind a single
// endpoint, per the CWCLOCK_OTEL_ENDPOINT/CWCLOCK_OTEL_PROTO settings.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// ServiceName identifies this process in every span/log record it emits.
const ServiceName = "cwclock-api"

// The two supported values of CWCLOCK_OTEL_PROTO.
const (
	ProtoHTTP = "otlp/http"
)

// Config configures the single OTEL endpoint that receives traces and logs
// (and metrics, see internal/metrics), driven by CWCLOCK_OTEL_ENDPOINT/PROTO.
type Config struct {
	Endpoint string
	Proto    string
	Version  string
}

// Providers bundles what the rest of the app needs: a Tracer to start a span
// per request, a *slog.Logger set as the process default, and a Shutdown
// hook to flush pending spans/logs on exit.
type Providers struct {
	Tracer         trace.Tracer
	Logger         *slog.Logger
	TracerProvider *sdktrace.TracerProvider
	Shutdown       func(context.Context) error
}

// Setup always configures stdout/stderr logging. When cfg.Endpoint is set,
// it additionally exports traces and logs to that OTLP endpoint (gRPC by
// default, or HTTP when cfg.Proto is "otlp/http").
func Setup(ctx context.Context, cfg Config) (*Providers, error) {
	res := resource.NewSchemaless(
		attribute.String("service.name", ServiceName),
		attribute.String("service.version", cfg.Version),
	)

	handlers := []slog.Handler{newStdHandler(slog.LevelInfo)}
	tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	var shutdownFuncs []func(context.Context) error

	if strings.TrimSpace(cfg.Endpoint) != "" {
		traceExp, logExp, err := buildExporters(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("telemetry: building OTLP exporters: %w", err)
		}

		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExp),
		)

		loggerProvider := sdklog.NewLoggerProvider(
			sdklog.WithResource(res),
			sdklog.WithProcessor(sdklog.NewBatchProcessor(logExp)),
		)
		handlers = append(handlers, &otelSlogHandler{logger: loggerProvider.Logger(ServiceName)})
		shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown, loggerProvider.Shutdown)
	}

	otel.SetTracerProvider(tracerProvider)

	logger := slog.New(newMultiHandler(handlers))
	slog.SetDefault(logger)

	return &Providers{
		Tracer:         tracerProvider.Tracer(ServiceName),
		Logger:         logger,
		TracerProvider: tracerProvider,
		Shutdown: func(ctx context.Context) error {
			var errs []error
			for _, fn := range shutdownFuncs {
				if err := fn(ctx); err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("telemetry shutdown: %v", errs)
			}
			return nil
		},
	}, nil
}

func buildExporters(ctx context.Context, cfg Config) (sdktrace.SpanExporter, sdklog.Exporter, error) {
	if cfg.Proto == ProtoHTTP {
		traceExp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(cfg.Endpoint), otlptracehttp.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		logExp, err := otlploghttp.New(ctx, otlploghttp.WithEndpoint(cfg.Endpoint), otlploghttp.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
		return traceExp, logExp, nil
	}

	traceExp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(cfg.Endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	logExp, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpoint(cfg.Endpoint), otlploggrpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	return traceExp, logExp, nil
}
