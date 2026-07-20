import React, { useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FiCalendar, FiUpload } from "react-icons/fi";
import styles from "./Styles/TaskInput.module.css";
import useTimer from "./useTimer";
import useDateHook from "./useDateHook";
import { postTasksApi, startTask } from "../../Redux/Tasks/Task.actions";
import useTime from "./useTime";
import projectLabel from "../common/projectLabel";
import AutocompleteSelect from "../common/AutocompleteSelect";
import Tooltip from "../common/Tooltip";
import padTimeString from "../common/padTimeString";
import { useI18n } from "../../i18n/I18nContext";

const TaskInput = ({ isAdminOrOwner, onImportClick }) => {
  const { t } = useI18n();
  const { timerOn, sec, min, hrs, handleTimer } = useTimer();
  const { hours, minutes, seconds } = useDateHook();
  const { hours2, minutes2, seconds2 } = useTime();
  const [name, setName] = useState("");
  const [projectId, setProjectId] = useState("");
  const [allDayDate, setAllDayDate] = useState("");
  const allDayInputRef = useRef(null);
  const { start } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);
  const dispatch = useDispatch();

  // handleAllDayDate creates an all-day time entry on the picked day for the
  // currently selected project, as soon as a date is chosen - no separate
  // "create" click needed (ai-instruct-60).
  const handleAllDayDate = (e) => {
    const day = e.target.value;
    setAllDayDate(day);
    if (!day) return;
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;
    const taskObj = {
      text: name || project.name,
      day,
      allDay: true,
      clientId: project.clientId,
      projectId: project.id,
    };
    dispatch(postTasksApi(taskObj, currentOrgId, user.token));
    setAllDayDate("");
  };

  // Opens the (visually hidden) native date input's picker from the icon
  // button - showPicker() is the modern, no-visible-input way to trigger it;
  // .click() is the fallback for browsers that don't support it yet.
  const openAllDayPicker = () => {
    const el = allDayInputRef.current;
    if (!el) return;
    if (el.showPicker) {
      el.showPicker();
    } else {
      el.click();
    }
  };

  const handleSubmit = () => {
    if (timerOn) {
      const project = projects.find((p) => p.id === projectId);
      if (!project) {
        return;
      }
      let taskObj = {
        text: name || project.name,
        day: new Date().toISOString().slice(0, 10),
        start: start,
        end: padTimeString(`${hours2}:${minutes2}:${seconds2}`),
        allDay: false,
        clientId: project.clientId,
        projectId: project.id,
      };
      dispatch(postTasksApi(taskObj, currentOrgId, user.token));
      handleTimer();
    } else {
      handleTimer();
      dispatch(startTask(padTimeString(`${hours}:${minutes}:${seconds}`)));
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
            <Tooltip label={t("timeTracker.importCsvTitle")} position="bottom">
              <button
                type="button"
                className={styles.AllDayIconBtn}
                onClick={onImportClick}
                aria-label={t("timeTracker.importCsvTitle")}
              >
                <FiUpload fontSize="16px" />
              </button>
            </Tooltip>
          )}
          <input
            ref={allDayInputRef}
            type="date"
            className={styles.AllDayPickerHidden}
            value={allDayDate}
            onChange={handleAllDayDate}
            disabled={!projectId || timerOn}
            aria-hidden="true"
            tabIndex={-1}
          />
          <Tooltip label={t("timeTracker.allDay")} position="bottom">
            <button
              type="button"
              className={styles.AllDayIconBtn}
              onClick={openAllDayPicker}
              disabled={!projectId || timerOn}
              aria-label={t("timeTracker.allDay")}
            >
              <FiCalendar fontSize="16px" />
            </button>
          </Tooltip>
        </div>
      </div>
    </div>
  );
};

export default TaskInput;
