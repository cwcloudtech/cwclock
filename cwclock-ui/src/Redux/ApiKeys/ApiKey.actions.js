import axios from "axios";
import { toast } from "react-toastify";
import {
  ApiKeyLOADING,
  ApiKeyERROR,
  ApiKeyListSUCCESS,
  ApiKeyCreateSUCCESS,
  ApiKeyDeleteSUCCESS,
} from "./ApiKey.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/me/api-keys/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listApiKeysApi = (token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    const { data } = await axios.get(ENDPOINT, authConfig(token));
    dispatch({ type: ApiKeyListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
  }
};

export const createApiKeyApi = (fields, token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    const { data } = await axios.post(ENDPOINT, fields, authConfig(token));
    dispatch({ type: ApiKeyCreateSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteApiKeyApi = (id, token) => async (dispatch) => {
  dispatch({ type: ApiKeyLOADING });
  try {
    await axios.delete(ENDPOINT + id, authConfig(token));
    dispatch({ type: ApiKeyDeleteSUCCESS, payload: id });
    toast.success(translate(getStoredLocale(), "toasts.apiKeyDeleted"), toastOptions);
  } catch (e) {
    dispatch({ type: ApiKeyERROR });
  }
};
