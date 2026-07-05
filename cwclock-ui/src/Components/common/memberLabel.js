// Renders a member's display name, falling back to their email when no
// name/surname has been set yet (eg. members who signed up before this field existed).
const memberLabel = (member) => {
  return member.name || member.surname ? `${member.name || ""} ${member.surname || ""}`.trim() : member.email;
}

export default memberLabel;
