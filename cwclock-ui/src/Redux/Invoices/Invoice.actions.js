import axios from "axios";
import { toast } from "react-toastify";
import { InvoiceLOADING, InvoiceERROR, InvoiceListSUCCESS, InvoiceUpdateSUCCESS } from "./Invoice.types";
import toastOptions from "../toastOptions";
import { apiErrorMessage, translate, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/invoices/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

const toPayload = (clientId, start, end) => ({
  clientId,
  dateRangeStart: `${start}T00:00:00.000Z`,
  dateRangeEnd: `${end}T23:59:59.999Z`,
});

// A blob-response error carries the JSON error body as a Blob instead of a
// parsed object, so apiErrorMessage can't read it directly; this recovers
// the real i18n_code/message before falling back to a generic error.
const parseBlobError = async (e) => {
  if (e.response?.data instanceof Blob) {
    try {
      e.response.data = JSON.parse(await e.response.data.text());
    } catch {
      // leave as-is; apiErrorMessage falls back to a generic message
    }
  }
  return e;
};

const filenameFromDisposition = (disposition, fallback) => {
  const match = /filename="?([^"]+)"?/.exec(disposition || "");
  return match ? match[1] : fallback;
};

const downloadBlob = (response, fallbackFilename) => {
  const filename = filenameFromDisposition(response.headers["content-disposition"], fallbackFilename);
  const url = window.URL.createObjectURL(response.data);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
};

export const listInvoicesApi = (orgId, clientId, start, end, token) => async (dispatch) => {
  dispatch({ type: InvoiceLOADING });
  try {
    const { data } = await axios.get(ENDPOINT(orgId), {
      ...authConfig(token),
      params: { clientId, start, end },
    });
    dispatch({ type: InvoiceListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: InvoiceERROR });
  }
};

// Preview only renders the PDF and downloads it - nothing is saved, so
// there's no list state to update afterward.
export const previewInvoiceApi = (orgId, clientId, start, end, token) => async () => {
  try {
    const response = await axios.post(`${ENDPOINT(orgId)}preview`, toPayload(clientId, start, end), {
      ...authConfig(token),
      responseType: "blob",
    });
    downloadBlob(response, "invoice-preview.pdf");
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

// Generate saves the invoice and downloads its PDF. The caller is
// responsible for re-dispatching listInvoicesApi afterward to refresh the
// table, since this endpoint streams a PDF rather than the saved invoice.
export const generateInvoiceApi = (orgId, clientId, start, end, token) => async () => {
  try {
    const response = await axios.post(ENDPOINT(orgId), toPayload(clientId, start, end), {
      ...authConfig(token),
      responseType: "blob",
    });
    downloadBlob(response, "invoice.pdf");
    toast.success(translate(getStoredLocale(), "toasts.invoiceGenerated"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

export const downloadInvoicePdfApi = (orgId, invoiceId, token) => async () => {
  try {
    const response = await axios.get(`${ENDPOINT(orgId)}${invoiceId}/pdf`, {
      ...authConfig(token),
      responseType: "blob",
    });
    downloadBlob(response, "invoice.pdf");
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

export const updateInvoiceStatusApi = (orgId, invoiceId, status, token) => async (dispatch) => {
  try {
    const { data } = await axios.put(`${ENDPOINT(orgId)}${invoiceId}`, { status }, authConfig(token));
    dispatch({ type: InvoiceUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.invoiceUpdated"), toastOptions);
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};
