package handlers

import (
	"cwclock/client"
	"cwclock/utils"
	"fmt"
	"strings"
)

var validExportTypes = []string{"summary", "detailed"}
var validExportFileFormats = []string{"pdf", "csv"}

// HandleExport downloads a summary or detailed time report as a PDF or CSV
// file, optionally filtered to one client - unlike an invoice, a report
// covers every client by default (see cwclock-api's exportRequest.Clients).
func HandleExport(orgOverride string, clientID string, beginExpr string, endExpr string, toExpr string, reportType string, fileFormat string, outputOverride string) error {
	normalizedType := strings.ToLower(strings.TrimSpace(reportType))
	if utils.IsBlank(normalizedType) {
		normalizedType = "summary"
	}
	if !utils.ContainsValue(normalizedType, validExportTypes) {
		return fmt.Errorf("invalid --type %q: expected summary or detailed", reportType)
	}

	normalizedFormat := strings.ToLower(strings.TrimSpace(fileFormat))
	if utils.IsBlank(normalizedFormat) {
		normalizedFormat = "pdf"
	}
	if !utils.ContainsValue(normalizedFormat, validExportFileFormats) {
		return fmt.Errorf("invalid --file-format %q: expected pdf or csv", fileFormat)
	}

	start, end, err := resolveDateRangeParams(beginExpr, endExpr, toExpr)
	if err != nil {
		return err
	}

	orgID, err := resolveOrgID(orgOverride)
	if err != nil {
		return err
	}

	cli, err := client.NewClient()
	if err != nil {
		return err
	}

	req := client.ReportRequest{
		ExportType:     strings.ToUpper(normalizedFormat),
		DateRangeStart: start,
		DateRangeEnd:   end,
	}
	if trimmedClientID := strings.TrimSpace(clientID); utils.IsNotBlank(trimmedClientID) {
		req.Clients = &client.ReportIDFilter{IDs: []string{trimmedClientID}}
	}

	data, filename, err := cli.GenerateReport(orgID, normalizedType, req)
	if err != nil {
		return err
	}

	fallback := fmt.Sprintf("%s.%s", normalizedType, normalizedFormat)
	savedPath, err := saveDownloadedFile(data, filename, outputOverride, fallback)
	if err != nil {
		return err
	}

	fmt.Printf("Export saved to %s\n", savedPath)
	return nil
}
