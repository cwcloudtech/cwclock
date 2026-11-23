import axios from "axios";
import { toast } from "react-toastify";
import {
  AdminLOADING,
  AdminERROR,
  AdminUsersListSUCCESS,
  AdminUserUpdateSUCCESS,
  AdminUserDeleteSUCCESS,
  AdminOrgsListSUCCESS,
  AdminAuditListSUCCESS,
} from "./Admin.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const USERS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/users/`;
const ORGS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/organizations/`;
const AUDIT_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/audit/`;
const IPBANS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/ip-bans/`;

// Page size for the server-paginated admin lists (startIndex is a 0-based page
// number, maxResults the page size), consumed by the infinite-scroll screens.
export const ADMIN_PAGE_SIZE = Number(process.env.REACT_APP_PAGE_SIZE) || 20;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

// listAllUsersApi loads one page of the user roster. term filters by
// name/surname/email/id (case-insensitive, a UUID id matches too), role filters
// by global role, and startIndex/maxResults page the result. The reducer
// appends when startIndex > 0 (infinite scroll) and replaces on page 0.
export const listAllUsersApi =
  (token, { term = "", role = "", startIndex = 0, maxResults = ADMIN_PAGE_SIZE } = {}) =>
  async (dispatch) => {
    dispatch({ type: AdminLOADING });
    try {
      const { data } = await axios.get(USERS_ENDPOINT, {
        ...authConfig(token),
        params: { term: term || undefined, role: role || undefined, startIndex, maxResults },
      });
      dispatch({ type: AdminUsersListSUCCESS, payload: data });
      return data;
    } catch (e) {
      dispatch({ type: AdminERROR });
    }
  };

export const updateUserApi = (id, fields, token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.put(USERS_ENDPOINT + id, fields, authConfig(token));
    dispatch({ type: AdminUserUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.userUpdated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
    throw e;
  }
};

export const deleteUserApi = (id, token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    await axios.delete(USERS_ENDPOINT + id, authConfig(token));
    dispatch({ type: AdminUserDeleteSUCCESS, payload: id });
    toast.success(translate(getStoredLocale(), "toasts.userDeleted"), toastOptions);
  } catch (e) {
    dispatch({ type: AdminERROR });
    throw e;
  }
};

// disableMfaApi turns off every MFA factor on an account (see
// ai-instruct-68), for when a user is locked out of their authenticator or
// security key.
export const disableMfaApi = (id, token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.post(`${USERS_ENDPOINT}${id}/disable-mfa`, {}, authConfig(token));
    dispatch({ type: AdminUserUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.mfaDisabled"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
    throw e;
  }
};

// listAllOrganizationsApi loads one page of the organization roster. term
// filters by name/owner-email/id (case-insensitive, a UUID id matches too), and
// startIndex/maxResults page the result. The reducer appends when
// startIndex > 0 (infinite scroll) and replaces on page 0.
export const listAllOrganizationsApi =
  (token, { term = "", startIndex = 0, maxResults = ADMIN_PAGE_SIZE } = {}) =>
  async (dispatch) => {
    dispatch({ type: AdminLOADING });
    try {
      const { data } = await axios.get(ORGS_ENDPOINT, {
        ...authConfig(token),
        params: { term: term || undefined, startIndex, maxResults },
      });
      dispatch({ type: AdminOrgsListSUCCESS, payload: data });
      return data;
    } catch (e) {
      dispatch({ type: AdminERROR });
    }
  };

// listAuditApi loads one page of the audit trail. The filters and the page
// are sent as query params and the backend decides what is visible, so a
// hand-edited URL can't widen what comes back - the route itself is
// superuser-only.
export const listAuditApi = ({ page, size, action, subjectType, subject, from, to }, token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.get(AUDIT_ENDPOINT, {
      ...authConfig(token),
      params: {
        page,
        size,
        action: action || undefined,
        subjectType: subjectType || undefined,
        subject: subject || undefined,
        from: from || undefined,
        to: to || undefined,
      },
    });
    dispatch({ type: AdminAuditListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
  }
};

// --- IP administration (superuser) ------------------------------------------
// These hold their result in the calling component's local state rather than in
// the admin redux slice (like the per-user MFA calls): the IP roster and a
// member's IP list are screen-local, not shared app state. They surface errors
// as a toast and rethrow so the component can stop its spinner.

// listIpBansApi loads one page of the address roster (every address seen, plus
// any banned before being seen), filtered by term and paged (0-based). Returns
// { results, page, pageSize, hasNext }.
export const listIpBansApi = (term, page, token) => async () => {
  try {
    const { data } = await axios.get(IPBANS_ENDPOINT, {
      ...authConfig(token),
      params: { term: term || undefined, page },
    });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// ipStatsApi returns { total, banned } for the badges above the roster.
export const ipStatsApi = (token) => async () => {
  const { data } = await axios.get(`${IPBANS_ENDPOINT}stats`, authConfig(token));
  return data;
};

// putIpBanApi bans or unbans an address (one endpoint, action selects the op).
export const putIpBanApi = ({ ip, action, motif }, token) => async () => {
  try {
    await axios.put(IPBANS_ENDPOINT, { ip, action, motif: motif || "" }, authConfig(token));
    toast.success(
      translate(getStoredLocale(), action === "unban" ? "toasts.ipUnbanned" : "toasts.ipBanned"),
      toastOptions,
    );
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// listUserIpsApi returns the addresses one account has connected from, newest
// first: [{ ip, count, firstSeen, lastSeen, banned, geo }].
export const listUserIpsApi = (id, token) => async () => {
  try {
    const { data } = await axios.get(`${USERS_ENDPOINT}${id}/ips`, authConfig(token));
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// --- Public holidays (superuser decision table, keyed by ISO country) --------
// These are plain API helpers (like the IP lookups): they return the data and
// don't touch the admin reducer, since the Holidays screen keeps its own list.

const HOLIDAYS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/holidays/`;

export const listHolidaysApi = (token, country) => async () => {
  const { data } = await axios.get(HOLIDAYS_ENDPOINT, {
    ...authConfig(token),
    params: country ? { country } : {},
  });
  return data;
};

export const createHolidayApi = (body, token) => async () => {
  const { data } = await axios.post(HOLIDAYS_ENDPOINT, body, authConfig(token));
  return data;
};

export const updateHolidayApi = (id, body, token) => async () => {
  const { data } = await axios.put(`${HOLIDAYS_ENDPOINT}${id}`, body, authConfig(token));
  return data;
};

export const deleteHolidayApi = (id, token) => async () => {
  await axios.delete(`${HOLIDAYS_ENDPOINT}${id}`, authConfig(token));
  return id;
};
