import {
  GetCalendarEntriesLOADING,
  GetCalendarEntriesSUCCESS,
  GetCalendarEntriesERROR,
  CreateCalendarEntrySUCCESS,
  UpdateCalendarEntrySUCCESS,
  DeleteCalendarEntrySUCCESS,
} from "./Calendar.types";

const initialState = {
  entries: [],
  isLoading: false,
  isError: false,
};

export const CalendarReducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case GetCalendarEntriesLOADING: {
      return { ...state, isLoading: true, isError: false };
    }
    case GetCalendarEntriesERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case GetCalendarEntriesSUCCESS: {
      const items = Array.isArray(payload?.items) ? payload.items : [];
      return { ...state, entries: items, isLoading: false };
    }
    case CreateCalendarEntrySUCCESS: {
      return { ...state, entries: [...state.entries, payload] };
    }
    case UpdateCalendarEntrySUCCESS: {
      return { ...state, entries: state.entries.map((e) => (e.id === payload.id ? payload : e)) };
    }
    case DeleteCalendarEntrySUCCESS: {
      return { ...state, entries: state.entries.filter((e) => e.id !== payload) };
    }
    default: {
      return state;
    }
  }
};
