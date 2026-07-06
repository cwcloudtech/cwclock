import axios from "axios";
import { toast } from "react-toastify";
import { ReportLOADING, ReportERROR, ReportSUCCESS } from "./Report.types";
import toastOptions from "../toastOptions";
import { apiErrorMessage, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/reports`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

const toParams = ({ type, start, end, clientIds, projectIds, userIds }) => {
  const params = { type, start, end };
  if (clientIds?.length) params.clientIds = clientIds.join(",");
  if (projectIds?.length) params.projectIds = projectIds.join(",");
  if (userIds?.length) params.userIds = userIds.join(",");
  return params;
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
    const { data } = await axios.get(ENDPOINT(orgId), { ...authConfig(token), params: toParams(filters) });
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
    const response = await axios.get(`${ENDPOINT(orgId)}/export`, {
      ...authConfig(token),
      params: { ...toParams(filters), format },
      responseType: "blob",
    });
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
