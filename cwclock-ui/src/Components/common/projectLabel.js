// Renders a project as "{customer} - {project}" for the project
// autocomplete, so typing filters by customer name first (it's the start of
// the string, which is what browsers match datalist options against), then
// by project name.
const projectLabel = (project, clients) => {
  const client = clients.find((c) => c.id === project.clientId);
  return `${client ? client.name : "?"} - ${project.name}`;
};

export default projectLabel;
