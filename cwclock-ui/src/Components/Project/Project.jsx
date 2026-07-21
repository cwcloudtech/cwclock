import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { Link } from "react-router-dom";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import styles from "./Styles/Project.module.css";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi, createProjectApi, deleteProjectApi } from "../../Redux/Projects/Project.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";
import MultiSelect from "../common/MultiSelect";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import EmptyState from "../common/EmptyState";
import NeedOrganizationEmptyState from "../common/NeedOrganizationEmptyState";
import CopyIdButton from "../common/CopyIdButton";
import EditProjectModal from "./EditProjectModal";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../common/permissions";
import { useI18n } from "../../i18n/I18nContext";

const initialFields = { clientId: "", name: "", color: "#1cb9f7", dailyRate: "", subdivisions: [] };

const Project = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);
  const { members } = useSelector((state) => state.organizations);
  const [fields, setFields] = useState(initialFields);
  const [editingProject, setEditingProject] = useState(null);
  const [deletingProject, setDeletingProject] = useState(null);
  const [clientFilter, setClientFilter] = useState([]);
  const [search, setSearch] = useState("");

  const isAdminOrOwner = computeIsAdminOrOwner(user, members);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const projectFormConfig = {
    name: "Project",
    fields: [
      {
        name: "clientId",
        type: "autocomplete",
        label: t("common.client"),
        required: true,
        placeholder: t("common.selectAClient"),
        options: clients.map((client) => ({ value: client.id, label: client.name })),
      },
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "color", type: "color", label: t("common.color") },
      { name: "dailyRate", type: "number", label: t("projects.dailyRate"), step: "0.01", min: "0" },
      {
        name: "subdivisions",
        type: "tags",
        label: t("projects.subdivisions"),
        placeholder: t("projects.subdivisionsPlaceholder"),
      },
    ],
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!fields.name || !fields.clientId || !currentOrgId) return;
    const dailyRate = fields.dailyRate === "" ? undefined : Number(fields.dailyRate);
    dispatch(createProjectApi(currentOrgId, fields.clientId, fields.name, fields.color, dailyRate, fields.subdivisions, user.token));
    setField("name", "");
  };

  const handleDelete = async () => {
    const projectId = deletingProject.id;
    setDeletingProject(null);
    await dispatch(deleteProjectApi(currentOrgId, projectId, user.token));
  };

  const clientFilterOptions = [...clients]
    .sort((a, b) => a.name.localeCompare(b.name))
    .map((c) => ({ value: c.id, label: c.name }));

  const visibleProjects = projects
    .filter((p) => clientFilter.length === 0 || clientFilter.includes(p.clientId))
    .filter((p) => p.name.toLowerCase().includes(search.trim().toLowerCase()))
    .sort((a, b) => a.name.localeCompare(b.name));

  if (!currentOrgId) {
    return <NeedOrganizationEmptyState body={t("organizations.selectOrCreateFirst")} />;
  }

  if (clients.length === 0) {
    return (
      <div className={styles.main}>
        <h1 className="cw-title">{t("projects.title")}</h1>
        <p>
          {t("projects.needClient")}{" "}
          <Link to="/dashboard/clients">{t("common.createOne")}</Link>.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("projects.title")}</h1>
      <CollapsiblePanel title={t("projects.createProject")}>
        <ConfigForm
          config={projectFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleSubmit}
          submitLabel={t("common.add")}
        />
      </CollapsiblePanel>

      {projects.length === 0 && (
        <EmptyState title={t("projects.emptyTitle")} body={t("projects.emptyBody")} />
      )}

      {projects.length > 0 && (
        <div className={styles.filters}>
          <MultiSelect
            label={t("common.client")}
            options={clientFilterOptions}
            selected={clientFilter}
            onChange={setClientFilter}
          />
          <input
            className={`cw-input ${styles.search}`}
            type="text"
            placeholder={t("common.search")}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      )}

      <ul className="cw-list">
        {visibleProjects.map((project) => {
          const client = clients.find((c) => c.id === project.clientId);
          return (
            <li className={`cw-list-item ${styles.projectItem}`} key={project.id}>
              <span className={styles.swatch} style={{ backgroundColor: project.color }} />
              <strong>{project.name}</strong>
              {client && <span className={styles.clientName}>{client.name}</span>}
              {isAdminOrOwner && project.dailyRate && (
                <span className={styles.clientName}>
                  {t("projects.dailyRateSet", { rate: project.dailyRate })}
                </span>
              )}
              <div className={styles.rowActions}>
                <CopyIdButton id={project.id} className={styles.iconBtn} />
                {isAdminOrOwner && (
                  <Tooltip label={t("common.edit")}>
                    <button
                      type="button"
                      className={styles.iconBtn}
                      onClick={() => setEditingProject(project)}
                    >
                      <FaRegEdit style={{ fontSize: "18px" }} />
                    </button>
                  </Tooltip>
                )}
                {isAdminOrOwner && (
                  <Tooltip label={t("common.delete")}>
                    <button
                      type="button"
                      className={`${styles.iconBtn} ${styles.iconBtnDanger}`}
                      onClick={() => setDeletingProject(project)}
                    >
                      <MdDeleteForever style={{ fontSize: "20px" }} />
                    </button>
                  </Tooltip>
                )}
              </div>
            </li>
          );
        })}
      </ul>

      <EditProjectModal
        show={!!editingProject}
        onClose={() => setEditingProject(null)}
        targetProject={editingProject}
        orgId={currentOrgId}
        token={user.token}
      />

      <ConfirmModal
        show={!!deletingProject}
        title={t("projects.deleteProjectTitle")}
        body={deletingProject ? t("projects.deleteProjectBody", { name: deletingProject.name }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleDelete}
        onCancel={() => setDeletingProject(null)}
      />
    </div>
  );
};

export default Project;
