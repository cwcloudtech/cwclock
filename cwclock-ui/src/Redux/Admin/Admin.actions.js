import axios from "axios";
import { AdminLOADING, AdminERROR, AdminUsersListSUCCESS, AdminUserUpdateSUCCESS } from "./Admin.types";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/admin/users/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listAllUsersApi = (token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.get(ENDPOINT, authConfig(token));
    dispatch({ type: AdminUsersListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
  }
};

export const updateUserApi = (id, fields, token) => async (dispatch) => {
  dispatch({ type: AdminLOADING });
  try {
    const { data } = await axios.put(ENDPOINT + id, fields, authConfig(token));
    dispatch({ type: AdminUserUpdateSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AdminERROR });
    throw e;
  }
};
