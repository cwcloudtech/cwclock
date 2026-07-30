import { fetchPdf } from "../../api/blobRequest";

// generateReportPdfApi POSTs the same {exportType:"PDF", dateRangeStart,
// dateRangeEnd} payload cwclock-ui's export scripts send (see
// cwclock-api/internal/handlers/report_handler.go's exportRequest),
// returning a local file path for PdfViewerScreen. "Simplified" per the
// instruction: just a date range and Summary vs Detailed, no
// client/project/user filters.
export const generateReportPdfApi = (orgId, reportType, { start, end }) => async () =>
  fetchPdf("POST", `/organizations/${orgId}/reports/${reportType}`, {
    exportType: "PDF",
    dateRangeStart: start,
    dateRangeEnd: end,
  });
