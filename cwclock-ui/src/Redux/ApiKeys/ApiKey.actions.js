import axios from "axios";
import { toast } from "react-toastify";
import {
  ApiKeyLOADING,
  ApiKeyERROR,
  ApiKeyListSUCCESS,
  ApiKeyCreateSUCCESS,
  ApiKeyDeleteSUCCESS,
} from "./ApiKey.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/me/api-keys/`;
const CONFIG_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/me/config`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

// The API key token only ever exists in the browser right after creation
// (see ApiKeyHandler.Create) - these two calls hand it to the backend via a
// header rather than persisting/looking it up, so it never touches the URL
// or server access logs.
const configRequest = (token, apiKeyToken, orgId) => ({
  headers: { Authorization: `Bearer ${token}`, "X-Config-Token": apiKeyToken },
  params: orgId ? { organization: orgId } : undefined,
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

export const listApiKeysApi = (token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    const { data } = await axios.get(ENDPOINT, authConfig(token));
    dispatch({ type: ApiKeyListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
  }
};

export const createApiKeyApi = (fields, token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    const { data } = await axios.post(ENDPOINT, fields, authConfig(token));
    dispatch({ type: ApiKeyCreateSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteApiKeyApi = (id, token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    await axios.delete(ENDPOINT + id, authConfig(token));
    dispatch({ type: ApiKeyDeleteSUCCESS, payload: id });
    toast.success(translate(getStoredLocale(), "toasts.apiKeyDeleted"), toastOptions);
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
  }
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

// Downloads the CLI config file (importable with `cwclock configure import`)
// for a just-created API key - apiKeyToken is the plaintext token, only
// ever available in the browser right after createApiKeyApi resolves.
export const downloadApiKeyConfigApi = (apiKeyToken, orgId, token) => async () => {
  try {
    const response = await axios.get(`${CONFIG_ENDPOINT}/file`, {
      ...configRequest(token, apiKeyToken, orgId),
      responseType: "blob",
    });
    downloadBlob(response, "config");
  } catch (e) {
    toast.error(apiErrorMessage(await parseBlobError(e), getStoredLocale()), toastOptions);
    throw e;
  }
};

// Fetches a QR code (as a data: URI) of the same config content, for a
// just-created API key - see downloadApiKeyConfigApi.
export const fetchApiKeyConfigQrApi = (apiKeyToken, orgId, token) => async () => {
  try {
    const { data } = await axios.get(`${CONFIG_ENDPOINT}/qr`, configRequest(token, apiKeyToken, orgId));
    return data.qrCodePng;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
