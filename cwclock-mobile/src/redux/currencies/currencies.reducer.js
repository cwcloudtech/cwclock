import { CurrenciesSUCCESS } from "./currencies.actions";

const initialState = { items: [] };

export const currenciesReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case CurrenciesSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
