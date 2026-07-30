// Ported from cwclock-ui/src/Components/common/projectLabel.js - renders a
// project as "{client} - {project}" for pickers.
const projectLabel = (project, clients) => {
  const client = clients.find((c) => c.id === project.clientId);
  return `${client ? client.name : "?"} - ${project.name}`;
};

export default projectLabel;
