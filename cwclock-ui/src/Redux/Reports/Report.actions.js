import axios from "axios";
import { toast } from "react-toastify";
import { ReportLOADING, ReportERROR, ReportSUCCESS } from "./Report.types";
import toastOptions from "../toastOptions";
import { apiErrorMessage, getStoredLocale } from "../../i18n/translate";

// filters.type ("summary"|"detailed") selects which endpoint to call, per
// the Clockify-style contract the backend now speaks: separate URLs for
// each report shape instead of a shared endpoint with a `type` query param.
const ENDPOINT = (orgId, type) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/reports/${type}`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

const toPayload = ({ start, end, clientIds, projectIds, userIds }) => {
  const payload = {
    dateRangeStart: `${start}T00:00:00.000Z`,
    dateRangeEnd: `${end}T23:59:59.999Z`,
  };
  if (clientIds?.length) payload.clients = { ids: clientIds };
  if (projectIds?.length) payload.projects = { ids: projectIds };
  if (userIds?.length) payload.users = { ids: userIds };
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

export const fetchReportApi = (orgId, filters, token) => async (dispatch) => {
  dispatch({ type: ReportLOADING });
  try {
    const { data } = await axios.post(ENDPOINT(orgId, filters.type), toPayload(filters), authConfig(token));
    dispatch({ type: ReportSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ReportERROR });
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

const filenameFromDisposition = (disposition, fallback) => {
  const match = /filename="?([^"]+)"?/.exec(disposition || "");
  return match ? match[1] : fallback;
};

export const exportReportApi = (orgId, filters, format, token) => async () => {
  try {
    const response = await axios.post(
      ENDPOINT(orgId, filters.type),
      { ...toPayload(filters), exportType: format.toUpperCase() },
      { ...authConfig(token), responseType: "blob" }
    );
    const filename = filenameFromDisposition(response.headers["content-disposition"], `report.${format}`);
    const url = window.URL.createObjectURL(response.data);
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};
