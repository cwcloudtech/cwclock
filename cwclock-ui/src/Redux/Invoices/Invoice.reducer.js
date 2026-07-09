import { InvoiceLOADING, InvoiceERROR, InvoiceListSUCCESS, InvoiceUpdateSUCCESS } from "./Invoice.types";

const initialState = {
  invoices: [],
  isLoading: false,
  isError: false,
};

export const InvoiceReducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case InvoiceLOADING: {
      return { ...state, isLoading: true };
    }
    case InvoiceERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case InvoiceListSUCCESS: {
      return { ...state, invoices: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    case InvoiceUpdateSUCCESS: {
      const invoices = state.invoices.map((i) => (i.id === payload.id ? payload : i));
      return { ...state, invoices };
    }
    default: {
      return state;
    }
  }
};
