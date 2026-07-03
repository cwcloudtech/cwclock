import { ProjectLOADING, ProjectERROR, ProjectListSUCCESS, ProjectCreateSUCCESS } from "./Project.types";

const initialstate = {
  projects: [],
  isLoading: false,
  isError: false,
};

export const ProjectReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case ProjectLOADING: {
      return { ...state, isLoading: true };
    }
    case ProjectERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case ProjectListSUCCESS: {
      return { ...state, projects: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    case ProjectCreateSUCCESS: {
      return { ...state, projects: [...state.projects, payload], isLoading: false };
    }
    default: {
      return state;
    }
  }
};
