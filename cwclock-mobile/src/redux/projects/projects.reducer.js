import { ProjectsSUCCESS, ProjectCreateSUCCESS, ProjectUpdateSUCCESS, ProjectDeleteSUCCESS } from "./projects.actions";

const initialState = { items: [] };

export const projectsReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case ProjectsSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    case ProjectCreateSUCCESS:
      return { ...state, items: [...state.items, payload] };
    case ProjectUpdateSUCCESS:
      return { ...state, items: state.items.map((p) => (p.id === payload.id ? payload : p)) };
    case ProjectDeleteSUCCESS:
      return { ...state, items: state.items.filter((p) => p.id !== payload) };
    default:
      return state;
  }
};
