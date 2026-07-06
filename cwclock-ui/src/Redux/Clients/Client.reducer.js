import {
  ClientLOADING,
  ClientERROR,
  ClientListSUCCESS,
  ClientCreateSUCCESS,
  ClientUpdateSUCCESS,
} from "./Client.types";

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
    case ClientUpdateSUCCESS: {
      const clients = state.clients.map((c) => (c.id === payload.id ? payload : c));
      return { ...state, clients, isLoading: false };
    }
    default: {
      return state;
    }
  }
};
