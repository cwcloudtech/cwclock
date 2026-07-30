import { OrganizationsLOADING, OrganizationsSUCCESS, OrganizationsERROR, MembersSUCCESS } from "./organizations.actions";

const initialState = {
  items: [],
  members: [],
  isLoading: false,
};

export const organizationsReducer = (state = initialState, { type, payload } = {}) => {
  switch (type) {
    case OrganizationsLOADING:
      return { ...state, isLoading: true };
    case OrganizationsERROR:
      return { ...state, isLoading: false };
    case OrganizationsSUCCESS:
      return { ...state, items: Array.isArray(payload) ? payload : [], isLoading: false };
    case MembersSUCCESS:
      return { ...state, members: Array.isArray(payload) ? payload : [] };
    default:
      return state;
  }
};
