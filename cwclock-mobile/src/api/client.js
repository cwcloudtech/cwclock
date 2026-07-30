import axios from "axios";

// session is kept in module scope (not React state) so the interceptor
// below always sees the latest value without every screen having to thread
// it through - src/redux/session/session.actions.js is the only place that
// calls setClientSession/clearClientSession, right after it persists the
// same values via src/storage/session.js.
let session = { apiUrl: "", apiKey: "", orgId: "" };

const client = axios.create();

client.interceptors.request.use((config) => {
  if (session.apiKey) {
    config.headers["X-Api-Key"] = session.apiKey;
  }
  return config;
});

// setClientSession points the shared axios instance at an org/API key -
// apiUrl is the bare host (e.g. "https://api.cwclock.me", no "/v1"), same
// convention as the CLI's api_url config value (cwclock-cli/httpcli:
// RequestPath does `hostname + "/v1" + path`).
export const setClientSession = (next) => {
  session = { ...session, ...next };
  client.defaults.baseURL = session.apiUrl ? `${session.apiUrl.replace(/\/+$/, "")}/v1` : undefined;
};

export const clearClientSession = () => {
  session = { apiUrl: "", apiKey: "", orgId: "" };
  client.defaults.baseURL = undefined;
};

export const getCurrentOrgId = () => session.orgId;

// getBaseUrl/getClientHeaders expose just enough of the current session for
// react-native-blob-util-based binary downloads (reports/invoices PDFs,
// src/redux/{reports,invoices}/*.actions.js) that bypass axios entirely -
// axios's XHR responseType handling isn't a good fit for large binary
// downloads straight to a file on RN.
export const getBaseUrl = () => client.defaults.baseURL;
export const getClientHeaders = () => (session.apiKey ? { "X-Api-Key": session.apiKey } : {});

export default client;
