import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaRegEdit, FaRegHourglass } from "react-icons/fa";
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

// Formats a remaining duration as a single decreasing unit - "3d", "2h",
// "1 min", "30s" - dropping to the next smaller unit once the current one
// hits zero, rather than a fixed-width "Xd Xh Xm Xs" breakdown, so it stays
// glanceable in a narrow list column.
const formatCountdown = (totalSeconds, nowLabel) => {
  if (totalSeconds <= 0) return nowLabel;
  if (totalSeconds >= 86400) return `${Math.floor(totalSeconds / 86400)}d`;
  if (totalSeconds >= 3600) return `${Math.floor(totalSeconds / 3600)}h`;
  if (totalSeconds >= 60) return `${Math.floor(totalSeconds / 60)} min`;
  return `${totalSeconds}s`;
};

// Seeds its countdown from nextRunInSeconds - the exact remaining duration
// the API computed from its own clock (see exportJobHandler.toExportJobResponse,
// ai-instruct-79) - then ticks it down by exactly one second locally.
// Diffing nextRunAt against the browser's own clock instead would drift
// whenever the client's clock isn't in sync with the server's; counting
// down a single server-provided duration sidesteps that entirely. Ticks its
// own interval rather than the whole row/list re-rendering on a shared one,
// and stops ticking on unmount.
const NextRunCell = ({ nextRunInSeconds }) => {
  const { t } = useI18n();
  const [remaining, setRemaining] = useState(nextRunInSeconds);

  useEffect(() => {
    setRemaining(nextRunInSeconds);
    if (nextRunInSeconds == null) return undefined;
    const id = setInterval(() => setRemaining((r) => (r == null ? r : Math.max(0, r - 1))), 1000);
    return () => clearInterval(id);
  }, [nextRunInSeconds]);

  if (remaining == null) {
    return <div className={styles.jobNextRun}>—</div>;
  }

  return (
    <div className={styles.jobNextRun}>
      <Tooltip label={t("exportJobs.nextRunTooltip")}>
        <FaRegHourglass className={styles.nextRunIcon} />
      </Tooltip>
      <span>{formatCountdown(remaining, t("exportJobs.nextRunNow"))}</span>
    </div>
  );
};

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
        <NextRunCell nextRunInSeconds={job.nextRunInSeconds} />
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
