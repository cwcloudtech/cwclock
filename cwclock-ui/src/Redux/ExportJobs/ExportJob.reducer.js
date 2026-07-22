import {
  ExportJobLOADING,
  ExportJobERROR,
  ExportJobListSUCCESS,
  ExportJobCreateSUCCESS,
  ExportJobUpdateSUCCESS,
  ExportJobDeleteSUCCESS,
} from "./ExportJob.types";

const initialState = {
  exportJobs: [],
  isLoading: false,
  error: null,
};

const ExportJobReducer = (state = initialState, action) => {
  switch (action.type) {
    case ExportJobLOADING:
      return {
        ...state,
        isLoading: true,
        error: null,
      };
    case ExportJobERROR:
      return {
        ...state,
        isLoading: false,
        error: action.payload,
      };
    case ExportJobListSUCCESS:
      return {
        ...state,
        exportJobs: action.payload,
        isLoading: false,
        error: null,
      };
    case ExportJobCreateSUCCESS:
      return {
        ...state,
        exportJobs: [action.payload, ...state.exportJobs],
        isLoading: false,
        error: null,
      };
    case ExportJobUpdateSUCCESS:
      return {
        ...state,
        exportJobs: state.exportJobs.map((job) =>
          job.id === action.payload.id ? action.payload : job
        ),
        isLoading: false,
        error: null,
      };
    case ExportJobDeleteSUCCESS:
      return {
        ...state,
        exportJobs: state.exportJobs.filter((job) => job.id !== action.payload),
        isLoading: false,
        error: null,
      };
    default:
      return state;
  }
};

export default ExportJobReducer;
