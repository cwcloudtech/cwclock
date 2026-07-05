import axios from "axios";
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
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const updateOrgApi = (orgId, fields, token) => async (dispatch) => {
  dispatch({ type: OrgLOADING });
  try {
    const { data } = await axios.put(`${ENDPOINT}${orgId}`, fields, authConfig(token));
    dispatch({ type: OrgUpdateSUCCESS, payload: data });
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
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};

export const setMemberRateApi = (orgId, userId, dailyRate, currency, token) => async (dispatch) => {
  try {
    await axios.put(
      `${ENDPOINT}${orgId}/members/${userId}/rate`,
      { dailyRate, currency },
      authConfig(token)
    );
    dispatch(listMembersApi(orgId, token));
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
    return data;
  } catch (e) {
    dispatch({ type: OrgERROR });
    throw e;
  }
};
