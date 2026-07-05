import axios from "axios";
import {
  AdminLOADING,
  AdminERROR,
  AdminUsersListSUCCESS,
  AdminUserUpdateSUCCESS,
  AdminUserDeleteSUCCESS,
  AdminOrgsListSUCCESS,
} from "./Admin.types";

const USERS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/users/`;
const ORGS_ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/organizations/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listAllUsersApi = (token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.get(USERS_ENDPOINT, authConfig(token));
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
  } catch (e) {
    dispatch({ type: AdminERROR });
    throw e;
  }
};

export const listAllOrganizationsApi = (token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.get(ORGS_ENDPOINT, authConfig(token));
    dispatch({ type: AdminOrgsListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
  }
};
