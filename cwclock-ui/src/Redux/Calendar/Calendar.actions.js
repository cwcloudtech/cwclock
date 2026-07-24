import axios from "axios";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";
import {
  GetCalendarEntriesLOADING,
  GetCalendarEntriesSUCCESS,
  GetCalendarEntriesERROR,
  CreateCalendarEntrySUCCESS,
  UpdateCalendarEntrySUCCESS,
  DeleteCalendarEntrySUCCESS,
} from "./Calendar.types";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/time-entries/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

// getCalendarEntriesApi loads every time entry within [start, end] (inclusive
// "YYYY-MM-DD") in one unpaginated call, unlike the classic time tracker's
// getTasksApi which pages newest-first - the Calendar view instead needs a
// whole visible month/week grid at once, so it keeps its own Redux domain
// rather than reusing the tasks slice's paginated shape.
export const getCalendarEntriesApi = (orgId, token, start, end) => async (dispatch) => {
  dispatch({ type: GetCalendarEntriesLOADING });
  try {
    const { data } = await axios.get(ENDPOINT(orgId) + "range", {
      ...authConfig(token),
      params: { start, end },
    });
    dispatch({ type: GetCalendarEntriesSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: GetCalendarEntriesERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
  }
};

export const createCalendarEntryApi = (item, orgId, token) => async (dispatch) => {
  try {
    const { data } = await axios.post(ENDPOINT(orgId), item, authConfig(token));
    dispatch({ type: CreateCalendarEntrySUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.timeRecordCreated"), toastOptions);
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const updateCalendarEntryApi = (entry, orgId, token) => async (dispatch) => {
  try {
    const { id } = entry;
    const { data } = await axios.put(ENDPOINT(orgId) + id, entry, authConfig(token));
    dispatch({ type: UpdateCalendarEntrySUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.timeRecordUpdated"), toastOptions);
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteCalendarEntryApi = (id, orgId, token) => async (dispatch) => {
  try {
    await axios.delete(ENDPOINT(orgId) + id, authConfig(token));
    dispatch({ type: DeleteCalendarEntrySUCCESS, payload: id });
    toast.success(translate(getStoredLocale(), "toasts.timeRecordDeleted"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
