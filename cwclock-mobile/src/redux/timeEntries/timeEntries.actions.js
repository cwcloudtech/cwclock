import client from "../../api/client";

export const TimeEntriesLOADING = "timeEntries/Loading";
export const TimeEntriesLOADINGMORE = "timeEntries/LoadingMore";
export const TimeEntriesSUCCESS = "timeEntries/Success";
export const TimeEntriesAppendSUCCESS = "timeEntries/AppendSuccess";
export const TimeEntriesERROR = "timeEntries/Error";
export const TimeEntryCreateSUCCESS = "timeEntries/CreateSuccess";
export const TimeEntryUpdateSUCCESS = "timeEntries/UpdateSuccess";
export const TimeEntryDeleteSUCCESS = "timeEntries/DeleteSuccess";

const PAGE_SIZE = 50;

// Ported from cwclock-ui/src/Redux/Tasks/Task.actions.js's getTasksApi -
// page 1 replaces the list (initial load / pull-to-refresh), page > 1
// appends (infinite scroll), same as the web time tracker screen.
export const listTimeEntriesApi = (orgId, page = 1) => async (dispatch) => {
  dispatch({ type: page === 1 ? TimeEntriesLOADING : TimeEntriesLOADINGMORE });
  try {
    const { data } = await client.get(`/organizations/${orgId}/time-entries/`, {
      params: { page, pageSize: PAGE_SIZE },
    });
    dispatch({ type: page === 1 ? TimeEntriesSUCCESS : TimeEntriesAppendSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: TimeEntriesERROR });
    throw e;
  }
};

// createTimeEntryApi covers every creation path (timer stop, all-day) - the
// caller builds `fields` (text/day/start/end/allDay/half/clientId/projectId),
// same payload shape the web app's postTasksApi sends.
export const createTimeEntryApi = (orgId, fields) => async (dispatch) => {
  const { data } = await client.post(`/organizations/${orgId}/time-entries/`, fields);
  dispatch({ type: TimeEntryCreateSUCCESS, payload: data });
  return data;
};

export const updateTimeEntryApi = (orgId, entry) => async (dispatch) => {
  const { data } = await client.put(`/organizations/${orgId}/time-entries/${entry.id}`, entry);
  dispatch({ type: TimeEntryUpdateSUCCESS, payload: data });
  return data;
};

export const deleteTimeEntryApi = (orgId, id) => async (dispatch) => {
  await client.delete(`/organizations/${orgId}/time-entries/${id}`);
  dispatch({ type: TimeEntryDeleteSUCCESS, payload: id });
};
