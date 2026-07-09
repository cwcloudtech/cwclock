import {
  ProjectLOADING,
  ProjectERROR,
  ProjectListSUCCESS,
  ProjectCreateSUCCESS,
  ProjectUpdateSUCCESS,
  ProjectDeleteSUCCESS,
} from "./Project.types";

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
    case ProjectUpdateSUCCESS: {
      const projects = state.projects.map((p) => (p.id === payload.id ? payload : p));
      return { ...state, projects, isLoading: false };
    }
    case ProjectDeleteSUCCESS: {
      const projects = state.projects.filter((p) => p.id !== payload);
      return { ...state, projects, isLoading: false };
    }
    default: {
      return state;
    }
  }
};
