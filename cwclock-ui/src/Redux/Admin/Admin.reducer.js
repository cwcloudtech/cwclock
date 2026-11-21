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
  // users / organizations accumulate across infinite-scroll pages; *Total is
  // the full count behind the current filter, so the screen knows when to stop
  // asking for more (loaded length < total ⇒ a next page exists).
  users: [],
  usersTotal: 0,
  organizations: [],
  organizationsTotal: 0,
  // The audit trail is paged server-side, so the slice holds one page at a
  // time plus what the screen's controls need: the actions present in the
  // whole trail (for the filter) and whether a next page exists.
  audit: { entries: [], actions: [], subjectTypes: [], hasMore: false },
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
      const results = Array.isArray(payload?.results) ? payload.results : [];
      const startIndex = payload?.startIndex ?? 0;
      const users = startIndex > 0 ? [...state.users, ...results] : results;
      return {
        ...state,
        users,
        usersTotal: payload?.totalResults ?? users.length,
        isLoading: false,
      };
    }
    case AdminUserUpdateSUCCESS: {
      const users = state.users.map((u) => (u.id === payload.id ? payload : u));
      return { ...state, users, isLoading: false };
    }
    case AdminUserDeleteSUCCESS: {
      const users = state.users.filter((u) => u.id !== payload);
      return { ...state, users, usersTotal: Math.max(0, state.usersTotal - 1), isLoading: false };
    }
    case AdminOrgsListSUCCESS: {
      const results = Array.isArray(payload?.results) ? payload.results : [];
      const startIndex = payload?.startIndex ?? 0;
      const organizations = startIndex > 0 ? [...state.organizations, ...results] : results;
      return {
        ...state,
        organizations,
        organizationsTotal: payload?.totalResults ?? organizations.length,
        isLoading: false,
      };
    }
    case AdminAuditListSUCCESS: {
      return {
        ...state,
        audit: {
          entries: Array.isArray(payload?.entries) ? payload.entries : [],
          actions: Array.isArray(payload?.actions) ? payload.actions : [],
          subjectTypes: Array.isArray(payload?.subjectTypes) ? payload.subjectTypes : [],
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
