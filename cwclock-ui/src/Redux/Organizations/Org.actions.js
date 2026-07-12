import axios from "axios";
import { toast } from "react-toastify";
import {
  OrgLOADING,
  OrgERROR,
  OrgListSUCCESS,
  OrgCreateSUCCESS,
  OrgSelect,
  OrgMembersSUCCESS,
  OrgOwnerTransferred,
  OrgUpdateSUCCESS,
  OrgDeleteSUCCESS,
} from "./Org.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/organizations/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listOrgsApi = (token) => async (dispatch) => {
  dispatch({ type: OrgLOADING });
  try {
    const { data } = await axios.get(ENDPOINT, authConfig(token));
    dispatch({ type: OrgListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
  }
};

export const createOrgApi = (fields, token) => async (dispatch) => {
  dispatch({ type: OrgLOADING });
  try {
    const { data } = await axios.post(ENDPOINT, fields, authConfig(token));
    dispatch({ type: OrgCreateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.orgCreated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const updateOrgApi = (orgId, fields, token) => async (dispatch) => {
  dispatch({ type: OrgLOADING });
  try {
    const { data } = await axios.patch(`${ENDPOINT}${orgId}`, fields, authConfig(token));
    dispatch({ type: OrgUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.orgUpdated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

// Adding a connection saves it immediately (ai-instruct-40) instead of
// requiring the whole organization form to be submitted, via a dedicated
// PATCH endpoint that only appends one connection.
export const addExternalConnectionApi = (orgId, connection, token) => async (dispatch) => {
  try {
    const { data } = await axios.patch(`${ENDPOINT}${orgId}/external-connections`, connection, authConfig(token));
    dispatch({ type: OrgUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.externalConnectionAdded"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

// Removing a connection also saves immediately, mirroring addExternalConnectionApi.
export const removeExternalConnectionApi = (orgId, connectionId, token) => async (dispatch) => {
  try {
    const { data } = await axios.patch(`${ENDPOINT}${orgId}/external-connections/${connectionId}`, {}, authConfig(token));
    dispatch({ type: OrgUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.externalConnectionRemoved"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const deleteOrgApi = (orgId, token) => async (dispatch) => {
  dispatch({ type: OrgLOADING });
  try {
    await axios.delete(`${ENDPOINT}${orgId}`, authConfig(token));
    dispatch({ type: OrgDeleteSUCCESS, payload: orgId });
    toast.success(translate(getStoredLocale(), "toasts.orgDeleted"), toastOptions);
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const selectOrg = (orgId) => (dispatch) => {
  localStorage.setItem("currentOrgId", orgId);
  dispatch({ type: OrgSelect, payload: orgId });
};

export const listMembersApi = (orgId, token) => async (dispatch) => {
  try {
    const { data } = await axios.get(`${ENDPOINT}${orgId}/members/`, authConfig(token));
    dispatch({ type: OrgMembersSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
  }
};

export const addMemberApi = (orgId, email, role, token) => async (dispatch) => {
  try {
    await axios.post(`${ENDPOINT}${orgId}/members/`, { email, role }, authConfig(token));
    dispatch(listMembersApi(orgId, token));
    toast.success(translate(getStoredLocale(), "toasts.memberAdded"), toastOptions);
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const removeMemberApi = (orgId, userId, token) => async (dispatch) => {
  try {
    await axios.delete(`${ENDPOINT}${orgId}/members/${userId}`, authConfig(token));
    dispatch(listMembersApi(orgId, token));
    toast.success(translate(getStoredLocale(), "toasts.memberRemoved"), toastOptions);
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const setMemberRateApi = (orgId, userId, dailyRate, token) => async (dispatch) => {
  try {
    await axios.put(
      `${ENDPOINT}${orgId}/members/${userId}/rate`,
      { dailyRate },
      authConfig(token)
    );
    dispatch(listMembersApi(orgId, token));
    toast.success(translate(getStoredLocale(), "toasts.dailyRateUpdated"), toastOptions);
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const transferOwnershipApi = (orgId, email, token) => async (dispatch) => {
  try {
    const { data } = await axios.put(`${ENDPOINT}${orgId}/owner`, { email }, authConfig(token));
    dispatch({ type: OrgOwnerTransferred, payload: data });
    dispatch(listMembersApi(orgId, token));
    toast.success(translate(getStoredLocale(), "toasts.ownershipTransferred"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};
