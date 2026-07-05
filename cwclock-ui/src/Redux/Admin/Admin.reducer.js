import { AdminLOADING, AdminERROR, AdminUsersListSUCCESS, AdminUserUpdateSUCCESS } from "./Admin.types";

const initialstate = {
  users: [],
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
    default: {
      return state;
    }
  }
};
