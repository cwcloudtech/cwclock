import {
  ApiKeyLOADING,
  ApiKeyERROR,
  ApiKeyListSUCCESS,
  ApiKeyCreateSUCCESS,
  ApiKeyDeleteSUCCESS,
} from "./ApiKey.types";

const initialstate = {
  keys: [],
  createdKey: null,
  isLoading: false,
  isError: false,
};

export const ApiKeyReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case ApiKeyLOADING: {
      return { ...state, isLoading: true };
    }
    case ApiKeyERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case ApiKeyListSUCCESS: {
      return { ...state, keys: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    case ApiKeyCreateSUCCESS: {
      return { ...state, keys: [payload, ...state.keys], createdKey: payload, isLoading: false };
    }
    case ApiKeyDeleteSUCCESS: {
      const keys = state.keys.filter((k) => k.id !== payload);
      return { ...state, keys, isLoading: false };
    }
    default: {
      return state;
    }
  }
};
