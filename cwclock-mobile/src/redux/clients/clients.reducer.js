import { ClientsSUCCESS } from "./clients.actions";

const initialState = { items: [] };

export const clientsReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case ClientsSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
