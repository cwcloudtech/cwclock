import { CountriesSUCCESS } from "./countries.actions";

const initialState = { items: [] };

export const countriesReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case CountriesSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
