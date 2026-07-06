import axios from "axios";
import { toast } from "react-toastify";
import {
  ClientLOADING,
  ClientERROR,
  ClientListSUCCESS,
  ClientCreateSUCCESS,
  ClientUpdateSUCCESS,
} from "./Client.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale } from "../../i18n/translate";

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
    toast.success(translate(getStoredLocale(), "toasts.clientCreated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ClientERROR });
    throw e;
  }
};

export const updateClientApi = (orgId, clientId, fields, token) => async (dispatch) => {
  dispatch({ type: ClientLOADING });
  try {
    const { data } = await axios.put(
      `${ENDPOINT}${orgId}/clients/${clientId}`,
      fields,
      authConfig(token)
    );
    dispatch({ type: ClientUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.clientUpdated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ClientERROR });
    throw e;
  }
};
