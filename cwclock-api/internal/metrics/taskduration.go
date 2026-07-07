package metrics

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"cwclock-api/internal/models"
	"cwclock-api/internal/report"
	"cwclock-api/internal/store"
)

// taskDurationProducer implements metric.Producer: on every collection (be
// it a /metrics scrape or a periodic OTLP push), it re-queries the last 24h
// of time entries and emits one gauge metric per distinct task name — the
// name itself carries the (sanitized) task name, like the summary report
// groups rows by task name/label, with the org/client/project/user as
// attributes on each data point.
type taskDurationProducer struct {
	timeEntries *store.TimeEntryStore
	clients     *store.ClientStore
}

type taskGroupKey struct {
	orgID, clientID, projectID, userID string
}

func (p *taskDurationProducer) Produce(ctx context.Context) ([]metricdata.ScopeMetrics, error) {
	entries, err := p.timeEntries.ListRecent(ctx)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}

	clientList, err := p.clients.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	clientByID := make(map[string]models.Client, len(clientList))
	for _, c := range clientList {
		clientByID[c.ID] = c
	}

	// metric name -> group key -> total duration (seconds)
	byMetric := map[string]map[taskGroupKey]float64{}

	for _, e := range entries {
		durSecs := report.DurationSecs(e, clientByID[e.ClientID])
		name := "cwclock_task_duration_seconds_" + SanitizeMetricName(e.Text)
		key := taskGroupKey{orgID: e.OrganizationID, clientID: e.ClientID, projectID: e.ProjectID, userID: e.UserID}

		groups, ok := byMetric[name]
		if !ok {
			groups = map[taskGroupKey]float64{}
			byMetric[name] = groups
		}
		groups[key] += float64(durSecs)
	}

	now := time.Now()
	metrics := make([]metricdata.Metrics, 0, len(byMetric))
	for name, groups := range byMetric {
		dataPoints := make([]metricdata.DataPoint[float64], 0, len(groups))
		for key, total := range groups {
			dataPoints = append(dataPoints, metricdata.DataPoint[float64]{
				Attributes: attribute.NewSet(
					attribute.String("user_id", key.userID),
					attribute.String("organization_id", key.orgID),
					attribute.String("client_id", key.clientID),
					attribute.String("project_id", key.projectID),
				),
				Time:  now,
				Value: total,
			})
		}
		metrics = append(metrics, metricdata.Metrics{
			Name:        name,
			Description: "Total task duration (seconds) in the last 24h",
			Unit:        "s",
			Data:        metricdata.Gauge[float64]{DataPoints: dataPoints},
		})
	}

	return []metricdata.ScopeMetrics{
		{
			Scope:   instrumentation.Scope{Name: meterName},
			Metrics: metrics,
		},
	}, nil
}
