import { ClientsSUCCESS, ClientCreateSUCCESS, ClientUpdateSUCCESS, ClientDeleteSUCCESS } from "./clients.actions";

const initialState = { items: [] };

export const clientsReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case ClientsSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    case ClientCreateSUCCESS:
      return { ...state, items: [...state.items, payload] };
    case ClientUpdateSUCCESS:
      return { ...state, items: state.items.map((c) => (c.id === payload.id ? payload : c)) };
    case ClientDeleteSUCCESS:
      return { ...state, items: state.items.filter((c) => c.id !== payload) };
    default:
      return state;
  }
};
