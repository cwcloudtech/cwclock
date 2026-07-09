// A global superuser is granted an implicit owner role in every
// organization by the backend (see cwclock-api's OrgMembership middleware),
// even ones they never joined, so they can manage or delete any
// organization's clients/projects/members. These helpers mirror that on the
// frontend so admin/owner-only actions stay usable for them too, instead of
// disappearing whenever the superuser isn't an actual member.
export const isSuperadmin = (user) => user?.role === "superuser";

export const memberRole = (user, members) => members.find((m) => m.userId === user.id)?.role;

export const isAdminOrOwner = (user, members) => {
  if (isSuperadmin(user)) return true;
  const role = memberRole(user, members);
  return role === "admin" || role === "owner";
};

export const isOrgOwner = (user, org) => isSuperadmin(user) || org?.ownerId === user.id;
