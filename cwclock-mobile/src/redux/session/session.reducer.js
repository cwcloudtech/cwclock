import {
  SessionRestoring,
  SessionRestored,
  SessionMissing,
  SessionNeedsOrg,
  SessionConnected,
  SessionCleared,
} from "./session.actions";

// status: "restoring" (booting up, checking storage) | "missing" (no/invalid
// session, show onboarding) | "needsOrg" (valid key, no org picked yet) |
// "connected" (ready for the main tab navigator).
const initialState = {
  status: "restoring",
  apiUrl: null,
  orgId: null,
  user: null,
};

export const sessionReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case SessionRestoring:
      return { ...state, status: "restoring" };
    case SessionMissing:
      return { ...initialState, status: "missing" };
    case SessionRestored:
      return { ...state, status: "connected", apiUrl: payload.apiUrl, orgId: payload.orgId, user: payload.user };
    case SessionNeedsOrg:
      return { ...state, status: "needsOrg", apiUrl: payload.apiUrl, orgId: null, user: payload.user };
    case SessionConnected:
      return { ...state, status: "connected", ...payload };
    case SessionCleared:
      return { ...initialState, status: "missing" };
    default:
      return state;
  }
};
