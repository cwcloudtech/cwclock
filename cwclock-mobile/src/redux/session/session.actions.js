import client, { setClientSession, clearClientSession } from "../../api/client";
import { saveSession, saveOrgId, loadSession, clearSession } from "../../storage/session";

export const SessionRestoring = "session/Restoring";
export const SessionRestored = "session/Restored";
export const SessionMissing = "session/Missing";
export const SessionNeedsOrg = "session/NeedsOrg";
export const SessionConnected = "session/Connected";
export const SessionCleared = "session/Cleared";

// restoreSessionApi runs once at app boot: if a session was saved from a
// previous run, wire the shared API client back up and confirm the API key
// is still accepted server-side - a revoked/deleted key must send the user
// back to onboarding, not into the tab navigator where every screen would
// just 401.
export const restoreSessionApi = () => async (dispatch) => {
  dispatch({ type: SessionRestoring });
  const saved = await loadSession();
  if (!saved) {
    dispatch({ type: SessionMissing });
    return;
  }

  setClientSession(saved);
  try {
    const { data: user } = await client.get("/users/me");
    dispatch({ type: SessionRestored, payload: { ...saved, user } });
  } catch (e) {
    clearClientSession();
    dispatch({ type: SessionMissing });
  }
};

// connectApi validates {apiUrl, apiKey} against GET /users/me - the one
// endpoint that accepts X-Api-Key without also requiring an org membership
// check - then either stores orgId as given (from a QR/pasted config with an
// org_id line) or reports back that none was resolved, so the caller
// (ConnectScreen/ScanQrScreen/ManualEntryScreen) routes to OrgPickerScreen.
// Mirrors the CLI's resolveOrgID fallback (cwclock-cli/handlers/record.go).
export const connectApi = ({ apiUrl, apiKey, orgId }) => async (dispatch) => {
  setClientSession({ apiUrl, apiKey, orgId: orgId || "" });
  const { data: user } = await client.get("/users/me");

  if (orgId) {
    await saveSession({ apiUrl, apiKey, orgId });
    dispatch({ type: SessionConnected, payload: { apiUrl, orgId, user } });
  } else {
    await saveSession({ apiUrl, apiKey, orgId: null });
    dispatch({ type: SessionNeedsOrg, payload: { apiUrl, user } });
  }
  return user;
};

export const selectOrgApi = (orgId) => async (dispatch) => {
  await saveOrgId(orgId);
  setClientSession({ orgId });
  dispatch({ type: SessionConnected, payload: { orgId } });
};

export const disconnectApi = () => async (dispatch) => {
  await clearSession();
  clearClientSession();
  dispatch({ type: SessionCleared });
};
