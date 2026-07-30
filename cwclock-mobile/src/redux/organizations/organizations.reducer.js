import {
  OrganizationsLOADING,
  OrganizationsSUCCESS,
  OrganizationsERROR,
  OrganizationUpdateSUCCESS,
  MembersSUCCESS,
  MemberAddSUCCESS,
  MemberUpdateSUCCESS,
  MemberRemoveSUCCESS,
} from "./organizations.actions";

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
    case OrganizationUpdateSUCCESS:
      return { ...state, items: state.items.map((o) => (o.id === payload.id ? payload : o)) };
    case MembersSUCCESS:
      return { ...state, members: Array.isArray(payload) ? payload : [] };
    case MemberAddSUCCESS:
      return { ...state, members: [...state.members, payload] };
    case MemberUpdateSUCCESS:
      return { ...state, members: state.members.map((m) => (m.userId === payload.userId ? payload : m)) };
    case MemberRemoveSUCCESS:
      return { ...state, members: state.members.filter((m) => m.userId !== payload) };
    default:
      return state;
  }
};
