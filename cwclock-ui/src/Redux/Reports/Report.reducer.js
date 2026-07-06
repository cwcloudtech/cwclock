import { ReportLOADING, ReportERROR, ReportSUCCESS } from "./Report.types";

const initialState = {
  data: null,
  isLoading: false,
  isError: false,
};

export const ReportReducer = (state = initialState, { type, payload }) => {
  switch (type) {
    case ReportLOADING: {
      return { ...state, isLoading: true, isError: false };
    }
    case ReportERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case ReportSUCCESS: {
      return { ...state, data: payload, isLoading: false, isError: false };
    }
    default: {
      return state;
    }
  }
};
