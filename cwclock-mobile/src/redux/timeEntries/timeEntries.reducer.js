import {
  TimeEntriesLOADING,
  TimeEntriesLOADINGMORE,
  TimeEntriesSUCCESS,
  TimeEntriesAppendSUCCESS,
  TimeEntriesERROR,
  TimeEntryCreateSUCCESS,
  TimeEntryUpdateSUCCESS,
  TimeEntryDeleteSUCCESS,
} from "./timeEntries.actions";

// Matches the API's own ORDER BY data->>'day' DESC, data->>'start' DESC, so
// an entry inserted/edited client-side (rather than re-fetched) lands in the
// same slot the server would put it in - ported from cwclock-ui's
// TaskReducer.js compareEntries.
const compareEntries = (a, b) => {
  if (a.day !== b.day) return a.day > b.day ? -1 : 1;
  const aStart = a.start || "";
  const bStart = b.start || "";
  if (aStart === bStart) return 0;
  return aStart > bStart ? -1 : 1;
};

const initialState = {
  items: [],
  isLoading: false,
  isLoadingMore: false,
  isError: false,
  page: 1,
  hasMore: false,
};

export const timeEntriesReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case TimeEntriesLOADING:
      return { ...state, isLoading: true, isError: false };
    case TimeEntriesLOADINGMORE:
      return { ...state, isLoadingMore: true };
    case TimeEntriesERROR:
      return { ...state, isLoading: false, isLoadingMore: false, isError: true };
    case TimeEntriesSUCCESS: {
      const items = Array.isArray(payload?.items) ? payload.items : [];
      return { ...state, items, isLoading: false, page: payload?.page ?? 1, hasMore: !!payload?.hasMore };
    }
    case TimeEntriesAppendSUCCESS: {
      const items = Array.isArray(payload?.items) ? payload.items : [];
      return {
        ...state,
        items: [...state.items, ...items],
        isLoadingMore: false,
        page: payload?.page ?? state.page,
        hasMore: !!payload?.hasMore,
      };
    }
    case TimeEntryCreateSUCCESS:
      return { ...state, items: [...state.items, payload].sort(compareEntries) };
    case TimeEntryUpdateSUCCESS:
      return { ...state, items: state.items.map((e) => (e.id === payload.id ? payload : e)).sort(compareEntries) };
    case TimeEntryDeleteSUCCESS:
      return { ...state, items: state.items.filter((e) => e.id !== payload) };
    default:
      return state;
  }
};
