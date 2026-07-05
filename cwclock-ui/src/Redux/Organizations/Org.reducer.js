import {
  OrgLOADING,
  OrgERROR,
  OrgListSUCCESS,
  OrgCreateSUCCESS,
  OrgSelect,
  OrgMembersSUCCESS,
  OrgOwnerTransferred,
  OrgUpdateSUCCESS,
  OrgDeleteSUCCESS,
} from "./Org.types";

const initialstate = {
  organizations: [],
  members: [],
  currentOrgId: localStorage.getItem("currentOrgId") || "",
  isLoading: false,
  isError: false,
};

export const OrgReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case OrgLOADING: {
      return { ...state, isLoading: true };
    }
    case OrgERROR: {
      return { ...state, isLoading: false, isError: true };
    }
    case OrgListSUCCESS: {
      const organizations = Array.isArray(payload) ? payload : [];
      const hasCurrent = organizations.some((o) => o.id === state.currentOrgId);
      const currentOrgId = hasCurrent ? state.currentOrgId : organizations[0]?.id || "";
      if (currentOrgId) {
        localStorage.setItem("currentOrgId", currentOrgId);
      }
      return { ...state, organizations, currentOrgId, isLoading: false };
    }
    case OrgCreateSUCCESS: {
      const organizations = [...state.organizations, payload];
      localStorage.setItem("currentOrgId", payload.id);
      return { ...state, organizations, currentOrgId: payload.id, isLoading: false };
    }
    case OrgSelect: {
      return { ...state, currentOrgId: payload };
    }
    case OrgMembersSUCCESS: {
      return { ...state, members: Array.isArray(payload) ? payload : [] };
    }
    case OrgOwnerTransferred:
    case OrgUpdateSUCCESS: {
      const organizations = state.organizations.map((o) => (o.id === payload.id ? payload : o));
      return { ...state, organizations, isLoading: false };
    }
    case OrgDeleteSUCCESS: {
      const organizations = state.organizations.filter((o) => o.id !== payload);
      const currentOrgId = state.currentOrgId === payload ? "" : state.currentOrgId;
      return { ...state, organizations, currentOrgId, isLoading: false };
    }
    default: {
      return state;
    }
  }
};
