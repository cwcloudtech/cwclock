import axios from "axios";
import { toast } from "react-toastify";
import {
  InvoiceLOADING,
  InvoiceERROR,
  InvoiceListSUCCESS,
  InvoiceUpdateSUCCESS,
  InvoiceDeleteSUCCESS,
} from "./Invoice.types";
import toastOptions from "../toastOptions";
import { apiErrorMessage, translate, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/invoices/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

const toPayload = (clientId, start, end, projectIds, number) => {
  const payload = {
    clientId,
    dateRangeStart: `${start}T00:00:00.000Z`,
    dateRangeEnd: `${end}T23:59:59.999Z`,
  };
  if (projectIds?.length) payload.projectIds = projectIds;
  if (number) payload.number = number;
  return payload;
};

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

// clearInvoices empties the invoice list, used when the requested date
// range is invalid so the page shows no stale result.
export const clearInvoices = () => (dispatch) => dispatch({ type: InvoiceListSUCCESS, payload: [] });

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
export const previewInvoiceApi = (orgId, clientId, start, end, projectIds, token) => async () => {
  try {
    const response = await axios.post(`${ENDPOINT(orgId)}preview`, toPayload(clientId, start, end, projectIds), {
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
// number is optional - when set, it's used as the invoice's id instead of
// the usual computed one (see "Generate with id").
export const generateInvoiceApi = (orgId, clientId, start, end, projectIds, token, number) => async () => {
  try {
    const response = await axios.post(ENDPOINT(orgId), toPayload(clientId, start, end, projectIds, number), {
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

// Reupload pushes an already-generated invoice's stored PDF to every one of
// the organization's external connections again (e.g. after fixing a
// connection's credentials). It doesn't change the invoice itself, so
// there's no store update to dispatch, only a success/error toast.
export const reuploadInvoiceApi = (orgId, invoiceId, token) => async () => {
  try {
    await axios.post(`${ENDPOINT(orgId)}${invoiceId}/reupload`, {}, authConfig(token));
    toast.success(translate(getStoredLocale(), "toasts.invoiceReuploaded"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

// SendEmail emails an already-generated invoice's PDF to the client's
// invoice recipients (or their plain email when that field is blank). It
// doesn't change the invoice itself, so there's no store update to
// dispatch, only a success/error toast.
export const sendInvoiceEmailApi = (orgId, invoiceId, token) => async () => {
  try {
    await axios.post(`${ENDPOINT(orgId)}${invoiceId}/send`, {}, authConfig(token));
    toast.success(translate(getStoredLocale(), "toasts.invoiceSent"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteInvoiceApi = (orgId, invoiceId, token) => async (dispatch) => {
  try {
    await axios.delete(`${ENDPOINT(orgId)}${invoiceId}`, authConfig(token));
    dispatch({ type: InvoiceDeleteSUCCESS, payload: invoiceId });
    toast.success(translate(getStoredLocale(), "toasts.invoiceDeleted"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};
