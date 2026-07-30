import client from "../../api/client";

export const OrganizationsLOADING = "organizations/Loading";
export const OrganizationsSUCCESS = "organizations/Success";
export const OrganizationsERROR = "organizations/Error";
export const OrganizationUpdateSUCCESS = "organizations/UpdateSuccess";
export const MembersSUCCESS = "organizations/MembersSuccess";
export const MemberAddSUCCESS = "organizations/MemberAddSuccess";
export const MemberUpdateSUCCESS = "organizations/MemberUpdateSuccess";
export const MemberRemoveSUCCESS = "organizations/MemberRemoveSuccess";

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

// listMembersApi loads the current organization's members - used both to
// resolve the connected user's own role (see common/permissions.js) and by
// MembersScreen's RBAC list (ai-instruct-106).
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

// updateOrganizationApi mirrors OrganizationHandler.Update - owner-only
// server-side (see route's RequireRole(RoleOwner)), gated the same way in
// OrganizationScreen. External connections (S3/Google Drive/Git sync
// targets) aren't part of the mobile form - the server ignores whatever a
// whole-org PATCH sends for that field anyway (see the handler's comment),
// keeping the org's existing connections untouched regardless.
export const updateOrganizationApi = (orgId, fields) => async (dispatch) => {
  const { data } = await client.patch(`/organizations/${orgId}`, fields);
  dispatch({ type: OrganizationUpdateSUCCESS, payload: data });
  return data;
};

// addMemberApi invites an existing cwclock user (by email) into the
// organization with a starting role - owner-only server-side. Mirrors
// cwclock-ui's Organizations.jsx addMemberApi call.
export const addMemberApi = (orgId, email, role) => async (dispatch) => {
  const { data } = await client.post(`/organizations/${orgId}/members/`, { email, role });
  dispatch({ type: MemberAddSUCCESS, payload: data });
  return data;
};

// updateMemberRoleApi changes an existing member's role - owner-only
// server-side. cwclock-ui has no equivalent screen (it only sets a role at
// invite time); this is the "user RBAC in the organization" feature
// ai-instruct-106 asks for on mobile specifically.
export const updateMemberRoleApi = (orgId, userId, role) => async (dispatch) => {
  const { data } = await client.put(`/organizations/${orgId}/members/${userId}`, { role });
  dispatch({ type: MemberUpdateSUCCESS, payload: data });
  return data;
};

export const removeMemberApi = (orgId, userId) => async (dispatch) => {
  await client.delete(`/organizations/${orgId}/members/${userId}`);
  dispatch({ type: MemberRemoveSUCCESS, payload: userId });
};
