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
