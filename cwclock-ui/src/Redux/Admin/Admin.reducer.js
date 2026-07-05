import {
  AdminLOADING,
  AdminERROR,
  AdminUsersListSUCCESS,
  AdminUserUpdateSUCCESS,
  AdminUserDeleteSUCCESS,
  AdminOrgsListSUCCESS,
} from "./Admin.types";

const initialstate = {
  users: [],
  organizations: [],
  isLoading: false,
  isError: false,
};

export const AdminReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case AdminLOADING: {
      return { ...state, isLoading: true };
    }
    case AdminERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case AdminUsersListSUCCESS: {
      return { ...state, users: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    case AdminUserUpdateSUCCESS: {
      const users = state.users.map((u) => (u.id === payload.id ? payload : u));
      return { ...state, users, isLoading: false };
    }
    case AdminUserDeleteSUCCESS: {
      const users = state.users.filter((u) => u.id !== payload);
      return { ...state, users, isLoading: false };
    }
    case AdminOrgsListSUCCESS: {
      return { ...state, organizations: Array.isArray(payload) ? payload : [], isLoading: false };
    }
    default: {
      return state;
    }
  }
};
