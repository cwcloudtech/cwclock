import client from "../../api/client";
import { fetchPdf } from "../../api/blobRequest";

export const InvoicesSUCCESS = "invoices/Success";
export const InvoicesERROR = "invoices/Error";

// listInvoicesApi mirrors InvoiceHandler.List - one client + date range,
// all three required by the API.
export const listInvoicesApi = (orgId, clientId, { start, end }) => async (dispatch) => {
  try {
    const { data } = await client.get(`/organizations/${orgId}/invoices/`, { params: { clientId, start, end } });
    dispatch({ type: InvoicesSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: InvoicesERROR });
    throw e;
  }
};

// previewInvoicePdfApi renders the PDF without saving anything (see
// InvoiceHandler.Preview) - safe to call repeatedly while adjusting the
// client/date range before committing to Generate.
export const previewInvoicePdfApi = (orgId, { clientId, start, end, projectIds }) => async () =>
  fetchPdf("POST", `/organizations/${orgId}/invoices/preview`, {
    clientId,
    dateRangeStart: start,
    dateRangeEnd: end,
    projectIds,
  });

// generateInvoicePdfApi allocates a real invoice number and saves it (see
// InvoiceHandler.Generate) - callers should confirm with the user first,
// since unlike Preview this can't be undone from the app.
export const generateInvoicePdfApi = (orgId, { clientId, start, end, projectIds }) => async () =>
  fetchPdf("POST", `/organizations/${orgId}/invoices`, {
    clientId,
    dateRangeStart: start,
    dateRangeEnd: end,
    projectIds,
  });

// downloadInvoicePdfApi fetches an already-generated invoice's stored PDF.
export const downloadInvoicePdfApi = (orgId, invoiceId) => async () =>
  fetchPdf("GET", `/organizations/${orgId}/invoices/${invoiceId}/pdf`);
