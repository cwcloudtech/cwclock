import { ClientLOADING, ClientERROR, ClientListSUCCESS, ClientCreateSUCCESS } from "./Client.types";

const initialstate = {
  clients: [],
  isLoading: false,
  isError: false,
};

export const ClientReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case ClientLOADING: {
      return { ...state, isLoading: true };
    }
    case ClientERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case ClientListSUCCESS: {
      return { ...state, clients: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    case ClientCreateSUCCESS: {
      return { ...state, clients: [...state.clients, payload], isLoading: false };
    }
    default: {
      return state;
    }
  }
};
