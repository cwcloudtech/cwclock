import RNBlobUtil from "react-native-blob-util";
import { getBaseUrl, getClientHeaders } from "./client";

// fetchPdf issues a binary request via react-native-blob-util (a better fit
// than axios for a large PDF response on RN, since it streams straight to a
// local file instead of round-tripping the whole body through the JS
// bridge) and returns that file's local path. On a non-2xx response it
// reads the (small) JSON error body for the same i18n_code/message shape
// the axios-based client's errors carry (see cwclock-api's writeError),
// attaching them to the thrown Error so i18n/translate.js's apiErrorMessage
// can format it the same way regardless of which request path failed.
export const fetchPdf = async (method, path, body) => {
  const res = await RNBlobUtil.config({ fileCache: true, appendExt: "pdf" }).fetch(
    method,
    `${getBaseUrl()}${path}`,
    { "Content-Type": "application/json", ...getClientHeaders() },
    body ? JSON.stringify(body) : undefined
  );

  const status = res.info().status;
  if (status >= 400) {
    let message = `Request failed (HTTP ${status})`;
    let i18nCode;
    try {
      const parsed = await res.json();
      i18nCode = parsed?.i18n_code;
      message = parsed?.message || message;
    } catch (e) {
      // Response wasn't JSON (unexpected for an error, but don't crash over it) - keep the generic message.
    }
    const err = new Error(message);
    err.i18nCode = i18nCode;
    throw err;
  }

  return res.path();
};
