import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import styles from "./Styles/TaskInput.module.css";
import useTimer from "./useTimer";
import useDateHook from "./useDateHook";
import { postTasksApi, startTask } from "../../Redux/Tasks/Task.actions";
import useTime from "./useTime";
import projectLabel from "../common/projectLabel";
import AutocompleteSelect from "../common/AutocompleteSelect";
import { useI18n } from "../../i18n/I18nContext";

const TaskInput = ({ isAdminOrOwner, onImportClick }) => {
  const { t } = useI18n();
  const { timerOn, sec, min, hrs, handleTimer } = useTimer();
  const { hours, minutes, seconds } = useDateHook();
  const { hours2, minutes2, seconds2 } = useTime();
  const [name, setName] = useState("");
  const [projectId, setProjectId] = useState("");
  const { start } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);
  const dispatch = useDispatch();

  const handleSubmit = () => {
    if (timerOn) {
      const project = projects.find((p) => p.id === projectId);
      if (!project) {
        return;
      }
      let taskObj = {
        text: name || t("timeTracker.defaultTaskName"),
        day: new Date().toISOString().slice(0, 10),
        start: start,
        end: `${hours2}:${minutes2}:${seconds2}`,
        allDay: false,
        clientId: project.clientId,
        projectId: project.id,
      };
      dispatch(postTasksApi(taskObj, currentOrgId, user.token));
      handleTimer();
    } else {
      handleTimer();
      dispatch(startTask(`${hours}:${minutes}:${seconds}`));
    }
  };

  return (
    <div className={styles.Task}>
      <div className={styles.Desc}>
        <input
          className={styles.textInput}
          type="text"
          placeholder={t("timeTracker.whatAreYouWorkingOn")}
          title={t("timeTracker.whatAreYouWorkingOn")}
          onChange={(e) => setName(e.target.value)}
        />
        <AutocompleteSelect
          className={styles.Projects}
          label={t("timeTracker.project")}
          placeholder={t("timeTracker.project")}
          options={projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) }))}
          value={projectId}
          onChange={setProjectId}
          disabled={timerOn}
        />
        <div className={styles.Timer}>
          <span className={styles.clock} title={t("timeTracker.elapsedTime")}>
            {hrs < 10 ? "0" + hrs : hrs}:{min < 10 ? "0" + min : min}:
            {sec < 10 ? "0" + sec : sec}
          </span>
          <button
            className={timerOn ? styles.Red : styles.Blue}
            onClick={handleSubmit}
            disabled={!projectId}
            title={timerOn ? t("timeTracker.stopTimer") : t("timeTracker.startTimer")}
          >
            {timerOn ? t("timeTracker.stop") : t("timeTracker.start")}
          </button>
          {isAdminOrOwner && (
            <button
              type="button"
              className={styles.ImportBtn}
              onClick={onImportClick}
              title={t("timeTracker.importCsvTitle")}
            >
              {t("timeTracker.importCsv")}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default TaskInput;
