// Ported from cwclock-ui/src/Components/common/permissions.js - a global
// superuser is granted an implicit owner role in every organization by the
// backend, so this mirrors that client-side rather than requiring an actual
// organization_members row.
export const isSuperadmin = (user) => user?.role === "superuser";

export const memberRole = (user, members) => members.find((m) => m.userId === user?.id)?.role;

export const isAdminOrOwner = (user, members) => {
  if (isSuperadmin(user)) return true;
  const role = memberRole(user, members);
  return role === "admin" || role === "owner";
};

// isOrgOwner gates owner-only actions (invite a member, change a member's
// role, edit the organization profile - see organizations.actions.js) -
// ownership is a property of the organization itself (ownerId), not the
// membership row, so it takes the org rather than the members list.
export const isOrgOwner = (user, org) => isSuperadmin(user) || org?.ownerId === user?.id;
