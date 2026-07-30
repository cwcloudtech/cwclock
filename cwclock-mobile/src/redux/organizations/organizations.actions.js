import client from "../../api/client";

export const OrganizationsLOADING = "organizations/Loading";
export const OrganizationsSUCCESS = "organizations/Success";
export const OrganizationsERROR = "organizations/Error";
export const MembersSUCCESS = "organizations/MembersSuccess";

// Ported from cwclock-ui/src/Redux/Organizations/Org.actions.js's
// listOrgsApi/listMembersApi - same endpoints and payload shapes, swapped
// from a Bearer token in authConfig to the shared client's X-Api-Key
// interceptor (src/api/client.js).
export const listOrganizationsApi = () => async (dispatch) => {
  dispatch({ type: OrganizationsLOADING });
  try {
    const { data } = await client.get("/organizations/");
    dispatch({ type: OrganizationsSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: OrganizationsERROR });
    throw e;
  }
};

// listMembersApi loads the current organization's members, used only to
// resolve the connected user's role there (see common/permissions.js) -
// mobile has no member management screen.
export const listMembersApi = (orgId) => async (dispatch) => {
  try {
    const { data } = await client.get(`/organizations/${orgId}/members/`);
    dispatch({ type: MembersSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: OrganizationsERROR });
    throw e;
  }
};
