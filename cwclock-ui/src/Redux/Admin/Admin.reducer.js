import {
  AdminLOADING,
  AdminERROR,
  AdminUsersListSUCCESS,
  AdminUserUpdateSUCCESS,
  AdminUserDeleteSUCCESS,
  AdminOrgsListSUCCESS,
  AdminAuditListSUCCESS,
} from "./Admin.types";

const initialstate = {
  users: [],
  organizations: [],
  // The audit trail is paged server-side, so the slice holds one page at a
  // time plus what the screen's controls need: the actions present in the
  // whole trail (for the filter) and whether a next page exists.
  audit: { entries: [], actions: [], hasMore: false },
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
    case AdminAuditListSUCCESS: {
      return {
        ...state,
        audit: {
          entries: Array.isArray(payload?.entries) ? payload.entries : [],
          actions: Array.isArray(payload?.actions) ? payload.actions : [],
          hasMore: !!payload?.hasMore,
        },
        isLoading: false,
      };
    }
    default: {
      return state;
    }
  }
};
