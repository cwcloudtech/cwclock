package client

import (
	"bufio"
	"bytes"
	"cwclock/utils"
	"fmt"
	"strings"
)

type MetricSample struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
	Value  string            `json:"value"`
}

type DisplayMetricSample struct {
	Name   string
	Labels string
	Value  string
}

func GetMetrics() ([]MetricSample, error) {
	return fetchAndParseMetrics("", "")
}

func GetMetricByName(name string, filter string) ([]MetricSample, error) {
	return fetchAndParseMetrics(name, filter)
}

func fetchAndParseMetrics(filterName string, filterLabel string) ([]MetricSample, error) {
	labelKey, labelValue, err := parseLabelFilter(filterLabel)
	if err != nil {
		return nil, err
	}

	cli, err := NewClient()
	if err != nil {
		return nil, err
	}

	body, err := cli.httpRequest(
		"/metrics",
		"GET",
		bytes.Buffer{},
		map[string]string{"Accept": "text/plain"},
	)
	if err != nil {
		return nil, err
	}

	defer body.Close()

	var samples []MetricSample
	scanner := bufio.NewScanner(body)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if utils.IsBlank(line) || strings.HasPrefix(line, "#") {
			continue
		}

		sample, ok := parseMetricLine(line)
		if !ok {
			continue
		}

		if utils.IsNotBlank(filterName) && !matchesMetricNameFilter(sample.Name, filterName) {
			continue
		}

		if utils.IsNotBlank(labelKey) {
			label, ok := sample.Labels[labelKey]
			if !ok || label != labelValue {
				continue
			}
		}

		samples = append(samples, sample)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return samples, nil
}

func matchesMetricNameFilter(metricName string, filterName string) bool {
	if strings.HasSuffix(filterName, "*") {
		prefix := strings.ToLower(strings.TrimSuffix(filterName, "*"))
		return strings.HasPrefix(strings.ToLower(metricName), prefix)
	}

	return strings.ToLower(metricName) == strings.ToLower(filterName)
}

func parseLabelFilter(filter string) (string, string, error) {
	if utils.IsBlank(filter) {
		return "", "", nil
	}

	trimmed := strings.TrimSpace(filter)
	sepIdx := strings.IndexAny(trimmed, ":=")
	if sepIdx < 0 {
		return "", "", fmt.Errorf("invalid filter format %q: expected label:value or label=value", filter)
	}

	key := strings.TrimSpace(trimmed[:sepIdx])
	value := strings.TrimSpace(trimmed[sepIdx+1:])
	if utils.IsBlank(key) || utils.IsBlank(value) {
		return "", "", fmt.Errorf("invalid filter format %q: expected label:value or label=value", filter)
	}

	return key, value, nil
}

func parseMetricLine(line string) (MetricSample, bool) {
	lastSpace := strings.LastIndex(line, " ")
	if lastSpace < 0 {
		return MetricSample{}, false
	}

	value := strings.TrimSpace(line[lastSpace+1:])
	rest := strings.TrimSpace(line[:lastSpace])

	labels := map[string]string{}
	name := rest

	if idx := strings.Index(rest, "{"); idx >= 0 {
		end := strings.LastIndex(rest, "}")
		if end < 0 {
			return MetricSample{}, false
		}
		name = rest[:idx]
		labelsStr := rest[idx+1 : end]
		labels = parsePrometheusLabels(labelsStr)
	}

	return MetricSample{
		Name:   name,
		Labels: labels,
		Value:  value,
	}, true
}

func parsePrometheusLabels(s string) map[string]string {
	result := map[string]string{}
	if utils.IsBlank(s) {
		return result
	}

	parts := strings.Split(strings.TrimSpace(s), ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		eqIdx := strings.Index(part, "=")
		if eqIdx < 0 {
			continue
		}

		k := strings.TrimSpace(part[:eqIdx])
		v := strings.TrimSpace(part[eqIdx+1:])
		v = strings.Trim(v, `"`)
		result[k] = v
	}

	return result
}
