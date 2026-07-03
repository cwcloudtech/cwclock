import axios from "axios";
import { ClientLOADING, ClientERROR, ClientListSUCCESS, ClientCreateSUCCESS } from "./Client.types";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/organizations/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listClientsApi = (orgId, token) => async (dispatch) => {
  dispatch({ type: ClientLOADING });
  try {
    const { data } = await axios.get(`${ENDPOINT}${orgId}/clients/`, authConfig(token));
    dispatch({ type: ClientListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ClientERROR });
  }
};

export const createClientApi = (orgId, fields, token) => async (dispatch) => {
  dispatch({ type: ClientLOADING });
  try {
    const { data } = await axios.post(
      `${ENDPOINT}${orgId}/clients/`,
      fields,
      authConfig(token)
    );
    dispatch({ type: ClientCreateSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ClientERROR });
    throw e;
  }
};
