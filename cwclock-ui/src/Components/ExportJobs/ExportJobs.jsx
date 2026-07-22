import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit } from "react-icons/fa";
import { MdDeleteForever } from "react-icons/md";
import { useI18n } from "../../i18n/I18nContext";
import Spinner from "../spinner/Spinner";
import EmptyState from "../common/EmptyState";
import NeedOrganizationEmptyState from "../common/NeedOrganizationEmptyState";
import Button from "../common/Button";
import Tooltip from "../common/Tooltip";
import ConfirmModal from "../common/ConfirmModal";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import {
  listExportJobsApi,
  createExportJobApi,
  updateExportJobApi,
  deleteExportJobApi,
} from "../../Redux/ExportJobs/ExportJob.actions";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../common/permissions";
import styles from "./Styles/ExportJobs.module.css";
import ExportJobModal from "./ExportJobModal";

const ExportJobRow = ({ job, orgId, token, onDelete, onEdit }) => {
  const { t } = useI18n();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const dispatch = useDispatch();

  const handleDelete = async () => {
    setShowDeleteConfirm(false);
    try {
      await dispatch(deleteExportJobApi(orgId, job.id, token));
      onDelete();
    } catch (e) {
      // error toast already shown by deleteExportJobApi
    }
  };

  return (
    <>
      <li className={`cw-list-item ${styles.jobRow}`}>
        <div className={styles.jobName}>{job.name}</div>
        <div className={styles.jobCron}>{job.cronExpression}</div>
        <div className={styles.jobReports}>{job.reportTypes.join(", ")}</div>
        <div className={styles.jobNextRun}>{job.nextRunAt ? new Date(job.nextRunAt).toLocaleString() : "—"}</div>
        <div className={styles.jobStatus}>
          {job.enabled ? (
            <span className={styles.statusEnabled}>{t("exportJobs.enabled")}</span>
          ) : (
            <span className={styles.statusDisabled}>{t("exportJobs.disabled")}</span>
          )}
        </div>
        <div className={styles.jobActions}>
          <Tooltip label={t("common.edit")} position="bottom">
            <button
              className={styles.actionBtn}
              onClick={() => onEdit(job)}
              title={t("common.edit")}
            >
              <FaRegEdit />
            </button>
          </Tooltip>
          <Tooltip label={t("common.delete")} position="bottom">
            <button
              className={styles.actionBtn}
              onClick={() => setShowDeleteConfirm(true)}
              title={t("common.delete")}
            >
              <MdDeleteForever />
            </button>
          </Tooltip>
        </div>
      </li>
      <ConfirmModal
        show={showDeleteConfirm}
        title={t("exportJobs.deleteConfirmTitle")}
        message={t("exportJobs.deleteConfirmMessage", { name: job.name })}
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </>
  );
};

const ExportJobs = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { exportJobs, isLoading } = useSelector((state) => state.exportJobs);
  const isAdminOrOwner = computeIsAdminOrOwner(user, members);
  const [showModal, setShowModal] = useState(false);
  const [editingJob, setEditingJob] = useState(null);

  useEffect(() => {
    if (user.token && currentOrgId && isAdminOrOwner) {
      dispatch(listExportJobsApi(currentOrgId, user.token));
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [user.token, currentOrgId, isAdminOrOwner, dispatch]);

  if (!isAdminOrOwner) {
    return (
      <div className={styles.container}>
        <NeedOrganizationEmptyState body={t("exportJobs.needAdminAccess")} />
      </div>
    );
  }

  if (!currentOrgId) {
    return (
      <div className={styles.container}>
        <NeedOrganizationEmptyState body={t("exportJobs.needOrganization")} />
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className={styles.container}>
        <Spinner />
      </div>
    );
  }

  const handleCreate = async (jobData) => {
    try {
      await dispatch(createExportJobApi(currentOrgId, jobData, user.token));
      setShowModal(false);
    } catch (e) {
      // error toast already shown by createExportJobApi
    }
  };

  const handleUpdate = async (jobData) => {
    try {
      await dispatch(updateExportJobApi(currentOrgId, editingJob.id, jobData, user.token));
      setShowModal(false);
      setEditingJob(null);
    } catch (e) {
      // error toast already shown by updateExportJobApi
    }
  };

  const handleEdit = (job) => {
    setEditingJob(job);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setEditingJob(null);
  };

  const openCreateModal = () => {
    setEditingJob(null);
    setShowModal(true);
  };

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>{t("exportJobs.title")}</h2>
        {exportJobs.length > 0 && (
          <Button onClick={openCreateModal} className={styles.createBtn}>
            {t("exportJobs.create")}
          </Button>
        )}
      </div>

      {exportJobs.length === 0 ? (
        <EmptyState
          icon="📊"
          title={t("exportJobs.emptyTitle")}
          body={t("exportJobs.emptyBody")}
          action={<Button onClick={openCreateModal}>{t("exportJobs.create")}</Button>}
        />
      ) : (
        <ul className={styles.jobsList}>
          {exportJobs.map((job) => (
            <ExportJobRow
              key={job.id}
              job={job}
              orgId={currentOrgId}
              token={user.token}
              onDelete={() => {
                dispatch(listExportJobsApi(currentOrgId, user.token));
              }}
              onEdit={handleEdit}
            />
          ))}
        </ul>
      )}

      <ExportJobModal
        show={showModal}
        job={editingJob}
        onSave={editingJob ? handleUpdate : handleCreate}
        onClose={handleCloseModal}
      />
    </div>
  );
};

export default ExportJobs;
