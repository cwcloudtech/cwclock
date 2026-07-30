import { InvoicesSUCCESS } from "./invoices.actions";

const initialState = { items: [] };

export const invoicesReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case InvoicesSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
