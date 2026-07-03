import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import styles from "./Styles/Project.module.css";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi, createProjectApi } from "../../Redux/Projects/Project.actions";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";

const initialFields = { clientId: "", name: "", color: "#1cb9f7" };

const Project = () => {
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);
  const [fields, setFields] = useState(initialFields);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const projectFormConfig = {
    name: "Project",
    fields: [
      {
        name: "clientId",
        type: "select",
        label: "Client",
        required: true,
        placeholder: "Select a client",
        options: clients.map((client) => ({ value: client.id, label: client.name })),
      },
      { name: "name", type: "text", label: "Name", required: true },
      { name: "color", type: "color", label: "Color" },
    ],
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!fields.name || !fields.clientId || !currentOrgId) return;
    dispatch(createProjectApi(currentOrgId, fields.clientId, fields.name, fields.color, user.token));
    setField("name", "");
  };

  if (!currentOrgId) {
    return <h1 className="cw-title">Select or create an organization first</h1>;
  }

  if (clients.length === 0) {
    return (
      <div className={styles.main}>
        <h1 className="cw-title">Projects</h1>
        <p>
          You need a client before you can create a project.{" "}
          <Link to="/dashboard/clients">Create one</Link>.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.main}>
      <h1 className="cw-title">Projects</h1>
      <CollapsiblePanel title="Create a project">
        <ConfigForm
          config={projectFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleSubmit}
          submitLabel="ADD"
        />
      </CollapsiblePanel>

      <ul className="cw-list">
        {projects.map((project) => {
          const client = clients.find((c) => c.id === project.clientId);
          return (
            <li className="cw-list-item" key={project.id}>
              <span style={{ color: project.color }}>{project.name}</span>{" "}
              {client ? `- ${client.name}` : ""}
            </li>
          );
        })}
      </ul>
    </div>
  );
};

export default Project;
