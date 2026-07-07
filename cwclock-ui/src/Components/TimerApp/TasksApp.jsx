import React, { useState } from "react";
import TaskInput from "./TaskInput";
import { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import TaskComponent from "./TaskComponent";
import Heading from "./Heading";
import styles from "./Styles/TaskApp.module.css";
import EmptyTask from "./EmptyTask";
import ImportCSVModal from "./ImportCSVModal";
import { useNavigate, Link } from "react-router-dom";
import Spinner from "../spinner/Spinner";
import { getTasksApi } from "../../Redux/Tasks/Task.actions";
import { listOrgsApi, listMembersApi } from "../../Redux/Organizations/Org.actions";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { useI18n } from "../../i18n/I18nContext";

const TasksApp = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [showImport, setShowImport] = useState(false);
  const { tasks, isLoading } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const myRole = members.find((m) => m.userId === user.id)?.role;
  const isAdminOrOwner = myRole === "admin" || myRole === "owner";

  useEffect(() => {
    if (user.token) {
      dispatch(listOrgsApi(user.token));
    } else {
      navigate("/login");
    }
  }, [user.token, dispatch, navigate]);

  useEffect(() => {
    if (user.token && currentOrgId) {
      dispatch(getTasksApi(currentOrgId, user.token));
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [user.token, currentOrgId, dispatch]);

  if (isLoading) {
    return (
      <div className={styles.Body1}>
        <Spinner />
      </div>
    );
  }

  if (!currentOrgId) {
    return (
      <div className={styles.Body1}>
        <div className={styles.Empty}>
          <p>
            {t("timeTracker.needOrganization")}{" "}
            <Link to="/dashboard/organizations">{t("common.createOne")}</Link>.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.Body1}>
      {isAdminOrOwner && (
        <div className={styles.toolbar}>
          <button
            type="button"
            className={styles.importBtn}
            onClick={() => setShowImport(true)}
            title={t("timeTracker.importCsvTitle")}
          >
            {t("timeTracker.importCsv")}
          </button>
        </div>
      )}
      <TaskInput />
      {tasks.length <= 0 ? (
        <div>
          <EmptyTask />
        </div>
      ) : (
        <div>
          <Heading />
          {tasks.map((item) => (
            <TaskComponent key={item.id} item={item} />
          ))}
        </div>
      )}
      <ImportCSVModal show={showImport} onClose={() => setShowImport(false)} />
    </div>
  );
};

export default TasksApp;
