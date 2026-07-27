package handlers

import (
	"cwclock/client"
	"cwclock/config"
	"cwclock/utils"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func HandleListMetrics(samples []client.MetricSample, formatOverride string) {
	if format := config.GetDefaultFormat(formatOverride); format == "json" {
		utils.PrintJson(samples)
	} else {
		displayMetricsAsTable(samples)
	}
}

func HandleGetMetric(name string, samples []client.MetricSample, formatOverride string, valueOnly bool) {
	if valueOnly {
		displayMetricValues(samples)
		return
	}

	if utils.IsEmpty(samples) {
		fmt.Printf("No metrics found for name: %s\n", name)
		return
	}

	if format := config.GetDefaultFormat(formatOverride); format == "json" {
		utils.PrintJson(samples)
	} else {
		displayMetricsAsTable(samples)
	}
}

func displayMetricValues(samples []client.MetricSample) {
	values := make([]string, 0, len(samples))
	for _, s := range samples {
		values = append(values, s.Value)
	}

	utils.PrintArray(values)
}

func toDisplaySamples(samples []client.MetricSample) []client.DisplayMetricSample {
	display := make([]client.DisplayMetricSample, 0, len(samples))
	for _, s := range samples {
		display = append(display, client.DisplayMetricSample{
			Name:   s.Name,
			Labels: formatMetricLabels(s.Labels),
			Value:  s.Value,
		})
	}
	return display
}

func formatMetricLabels(labels map[string]string) string {
	if utils.IsEmpty(labels) {
		return utils.EMPTY
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	return strings.Join(parts, ", ")
}

func displayMetricsAsTable(samples []client.MetricSample) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Labels", "Value"})
	table.SetAutoWrapText(false)
	table.SetColWidth(60)

	if utils.IsEmpty(samples) {
		table.Append([]string{"No metrics available", utils.EMPTY, utils.EMPTY})
		table.Render()
		return
	}

	for _, s := range samples {
		table.Append([]string{
			s.Name,
			formatMetricLabels(s.Labels),
			s.Value,
		})
	}
	table.Render()
}
