// Package metrics wires up OpenTelemetry metrics with two readers sharing
// the same instruments: a Prometheus exporter for GET /metrics, and (when
// configured) a periodic OTLP exporter pushing to the same single OTEL
// endpoint used for traces and logs.
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"cwclock-api/internal/middleware"
	"cwclock-api/internal/store"
	"cwclock-api/internal/telemetry"
	"cwclock-api/internal/utils"
)

const meterName = "cwclock-api"

// Config configures metrics export. Endpoint/Proto are the same
// CWCLOCK_OTEL_ENDPOINT/CWCLOCK_OTEL_PROTO values used for traces and logs.
type Config struct {
	Endpoint string
	Proto    string
	Version  string
}

// Metrics bundles the /metrics HTTP handler, the per-request observer to
// feed into the tracing middleware, the audit-event observer to hand to the
// audit store, and a shutdown hook.
type Metrics struct {
	Handler     http.Handler
	Observe     middleware.EndpointObserver
	RecordAudit store.AuditObserver
	Shutdown    func(context.Context) error
}

// Setup registers the default Go/process collectors, the custom business
// metrics (endpoint call counts/durations, user/client/project counts, task
// duration in the last 24h, audit events), and returns everything the
// router/main need.
func Setup(
	ctx context.Context,
	cfg Config,
	orgs *store.OrgStore,
	clients *store.ClientStore,
	projects *store.ProjectStore,
	timeEntries *store.TimeEntryStore,
) (*Metrics, error) {
	res := resource.NewSchemaless(
		attribute.String("service.name", telemetry.ServiceName),
		attribute.String("service.version", cfg.Version),
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	producer := &taskDurationProducer{timeEntries: timeEntries, clients: clients}

	promExporter, err := otelprometheus.New(
		otelprometheus.WithRegisterer(registry),
		otelprometheus.WithProducer(producer),
	)
	if err != nil {
		return nil, fmt.Errorf("metrics: building prometheus exporter: %w", err)
	}

	readerOpts := []sdkmetric.Option{sdkmetric.WithResource(res), sdkmetric.WithReader(promExporter)}
	var shutdownFuncs []func(context.Context) error

	if utils.IsNotBlank(cfg.Endpoint) {
		metricExp, err := buildExporter(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("metrics: building OTLP exporter: %w", err)
		}
		periodic := sdkmetric.NewPeriodicReader(metricExp, sdkmetric.WithProducer(producer))
		readerOpts = append(readerOpts, sdkmetric.WithReader(periodic))
		shutdownFuncs = append(shutdownFuncs, periodic.Shutdown)
	}

	mp := sdkmetric.NewMeterProvider(readerOpts...)
	otel.SetMeterProvider(mp)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)

	meter := mp.Meter(meterName)

	requestCounter, err := meter.Int64Counter(
		"http_server_requests_total",
		metric.WithDescription("Count of HTTP requests per endpoint"),
	)
	if err != nil {
		return nil, err
	}
	requestDuration, err := meter.Float64Histogram(
		"http_server_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests per endpoint"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	// Audit events, labelled exactly as the brief asks: by action, by tenant
	// and by user. Every one of those three is a bounded set on a CWClock
	// deployment - action is the fixed models.Audit* list, and the
	// organizations and accounts of a self-hosted time tracker number in the
	// tens, not the millions - so the cardinality of the series is bounded
	// too. That is the only reason a user id is acceptable as a label here;
	// on a deployment with a large or open signup, "user" would have to go
	// (an unbounded label is how a Prometheus server runs out of memory).
	auditCounter, err := meter.Int64Counter(
		"cwclock_audit_events_total",
		metric.WithDescription("Count of audit trail entries per action, organization and user"),
	)
	if err != nil {
		return nil, err
	}

	if err := registerGauges(meter, orgs, clients, projects); err != nil {
		return nil, err
	}

	observe := func(route, method string, status int, duration time.Duration) {
		attrs := metric.WithAttributes(
			attribute.String("route", route),
			attribute.String("method", method),
			attribute.Int("status", status),
		)
		requestCounter.Add(context.Background(), 1, attrs)
		requestDuration.Record(context.Background(), duration.Seconds(), attrs)
	}

	// An act with no organization (signing in) or no authenticated actor
	// (finishing a password reset) gets "none" rather than an empty label
	// value, so those series are legible in a query result instead of
	// looking like a scrape bug.
	recordAudit := func(action, organizationID, userID string) {
		auditCounter.Add(context.Background(), 1, metric.WithAttributes(
			attribute.String("action", action),
			attribute.String("organization", utils.If(utils.IsNotBlank(organizationID), organizationID, "none")),
			attribute.String("user", utils.If(utils.IsNotBlank(userID), userID, "none")),
		))
	}

	return &Metrics{
		Handler:     promhttp.HandlerFor(registry, promhttp.HandlerOpts{Registry: registry}),
		Observe:     observe,
		RecordAudit: recordAudit,
		Shutdown: func(ctx context.Context) error {
			var errs []error
			for _, fn := range shutdownFuncs {
				if err := fn(ctx); err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) > 0 {
				return fmt.Errorf("metrics shutdown: %v", errs)
			}
			return nil
		},
	}, nil
}

// registerGauges wires the "counter of users per role" and "counter of
// clients, projects" observable gauges to a single callback, so each scrape
// costs three cheap count queries rather than one per gauge.
func registerGauges(meter metric.Meter, orgs *store.OrgStore, clients *store.ClientStore, projects *store.ProjectStore) error {
	usersGauge, err := meter.Int64ObservableGauge(
		"cwclock_users_total",
		metric.WithDescription("Number of organization memberships per role"),
	)
	if err != nil {
		return err
	}
	clientsGauge, err := meter.Int64ObservableGauge(
		"cwclock_clients_total",
		metric.WithDescription("Number of clients"),
	)
	if err != nil {
		return err
	}
	projectsGauge, err := meter.Int64ObservableGauge(
		"cwclock_projects_total",
		metric.WithDescription("Number of projects"),
	)
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		if roleCounts, err := orgs.CountMembersByRole(ctx); err == nil {
			for role, count := range roleCounts {
				o.ObserveInt64(usersGauge, count, metric.WithAttributes(attribute.String("role", role)))
			}
		}
		if n, err := clients.Count(ctx); err == nil {
			o.ObserveInt64(clientsGauge, n)
		}
		if n, err := projects.Count(ctx); err == nil {
			o.ObserveInt64(projectsGauge, n)
		}
		return nil
	}, usersGauge, clientsGauge, projectsGauge)
	return err
}

func buildExporter(ctx context.Context, cfg Config) (sdkmetric.Exporter, error) {
	if cfg.Proto == telemetry.ProtoHTTP {
		return otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(cfg.Endpoint), otlpmetrichttp.WithInsecure())
	}
	return otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpoint(cfg.Endpoint), otlpmetricgrpc.WithInsecure())
}
