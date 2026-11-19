import axios from "axios";
import { toast } from "react-toastify";
import {
  AiLOADING,
  AiERROR,
  AiProvidersSUCCESS,
  AiConnectionsSUCCESS,
  AiConversationsSUCCESS,
  AiConversationSUCCESS,
  AiConversationRESET,
  AiSENDING,
  AiSendDONE,
} from "./AI.types";
import toastOptions from "../toastOptions";
import { getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const BASE = `${process.env.REACT_APP_APIURL}/v1/ai`;
const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

// --- providers + connections (subscriptions) ---

export const listProvidersApi = (token) => async (dispatch) => {
  try {
    const { data } = await axios.get(`${BASE}/providers`, authConfig(token));
    dispatch({ type: AiProvidersSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AiERROR });
  }
};

// listConnectionsApi loads the subscriptions the caller may spend. It never
// toasts on failure: it is also the call the sidebar gates the assistant on, so a
// hiccup should quietly leave the item hidden, not nag.
export const listConnectionsApi = (token) => async (dispatch) => {
  dispatch({ type: AiLOADING });
  try {
    const { data } = await axios.get(`${BASE}/connections`, authConfig(token));
    dispatch({ type: AiConnectionsSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AiERROR });
    return [];
  }
};

export const createConnectionApi = (fields, token) => async (dispatch) => {
  try {
    await axios.post(`${BASE}/connections`, fields, authConfig(token));
    const { data } = await axios.get(`${BASE}/connections`, authConfig(token));
    dispatch({ type: AiConnectionsSUCCESS, payload: data });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const updateConnectionApi = (id, fields, token) => async (dispatch) => {
  try {
    await axios.put(`${BASE}/connections/${id}`, fields, authConfig(token));
    const { data } = await axios.get(`${BASE}/connections`, authConfig(token));
    dispatch({ type: AiConnectionsSUCCESS, payload: data });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteConnectionApi = (id, token) => async (dispatch) => {
  try {
    await axios.delete(`${BASE}/connections/${id}`, authConfig(token));
    const { data } = await axios.get(`${BASE}/connections`, authConfig(token));
    dispatch({ type: AiConnectionsSUCCESS, payload: data });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// reorderConnectionsApi sets the order (the first is the default a question goes
// through when none is picked); used by the "set as default" action.
export const reorderConnectionsApi = (connectionIds, token) => async (dispatch) => {
  try {
    const { data } = await axios.post(`${BASE}/connections/order`, { connectionIds }, authConfig(token));
    dispatch({ type: AiConnectionsSUCCESS, payload: data });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// --- sharing ---

export const listSharesApi = (connectionId, token) => async () => {
  const { data } = await axios.get(`${BASE}/connections/${connectionId}/shares`, authConfig(token));
  return data;
};

export const shareConnectionApi = (connectionId, email, token) => async () => {
  try {
    const { data } = await axios.post(
      `${BASE}/connections/${connectionId}/shares`,
      { email, role: "user" },
      authConfig(token)
    );
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const unshareConnectionApi = (connectionId, userId, token) => async () => {
  try {
    await axios.delete(`${BASE}/connections/${connectionId}/shares/${userId}`, authConfig(token));
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// --- chat + conversations ---

export const listConversationsApi = (token) => async (dispatch) => {
  try {
    const { data } = await axios.get(`${BASE}/conversations`, authConfig(token));
    dispatch({ type: AiConversationsSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AiERROR });
    return [];
  }
};

export const getConversationApi = (id, token) => async (dispatch) => {
  dispatch({ type: AiLOADING });
  try {
    const { data } = await axios.get(`${BASE}/conversations/${id}`, authConfig(token));
    dispatch({ type: AiConversationSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: AiERROR });
    throw e;
  }
};

export const resetConversation = () => ({ type: AiConversationRESET });

// sendChatApi asks the assistant a question. conversationId continues a stored
// thread (blank starts one); connectionId picks the subscription (blank is the
// default); orgId tells the assistant which organization to target. Returns the
// server's answer so the chat can append it immediately.
export const sendChatApi = ({ message, conversationId, connectionId, orgId }, token) => async (dispatch) => {
  dispatch({ type: AiSENDING });
  try {
    const { data } = await axios.post(
      `${BASE}/chat`,
      {
        message,
        ...(conversationId ? { conversationId } : {}),
        ...(connectionId ? { connectionId } : {}),
        ...(orgId ? { orgId } : {}),
      },
      authConfig(token)
    );
    dispatch({ type: AiSendDONE });
    return data;
  } catch (e) {
    dispatch({ type: AiSendDONE });
    throw e;
  }
};

export const deleteConversationsApi = ({ ids, all }, token) => async (dispatch) => {
  try {
    await axios.delete(`${BASE}/conversations`, {
      ...authConfig(token),
      data: all ? { all: true } : { ids },
    });
    const { data } = await axios.get(`${BASE}/conversations`, authConfig(token));
    dispatch({ type: AiConversationsSUCCESS, payload: data });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
