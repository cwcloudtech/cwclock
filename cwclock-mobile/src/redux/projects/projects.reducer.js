import { ProjectsSUCCESS } from "./projects.actions";

const initialState = { items: [] };

export const projectsReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case ProjectsSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
