import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import styles from "./Styles/TaskInput.module.css";
import useTimer from "./useTimer";
import useDateHook from "./useDateHook";
import { postTasksApi, startTask } from "../../Redux/Tasks/Task.actions";
import useTime from "./useTime";

const TaskInput = () => {
  const { timerOn, sec, min, hrs, handleTimer } = useTimer();
  const { hours, minutes, seconds } = useDateHook();
  const { hours2, minutes2, seconds2 } = useTime();
  const [name, setName] = useState("");
  const [projectId, setProjectId] = useState("");
  const { start } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const dispatch = useDispatch();

  const handleSubmit = () => {
    if (timerOn) {
      const project = projects.find((p) => p.id === projectId);
      if (!project) {
        return;
      }
      let taskObj = {
        text: name || "Task",
        status: false,
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
          placeholder="What are you working on?"
          onChange={(e) => setName(e.target.value)}
        />
        <select
          className={styles.Projects}
          value={projectId}
          disabled={timerOn}
          onChange={(e) => setProjectId(e.target.value)}
        >
          <option value="">Project</option>
          {projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
        <div className={styles.Timer}>
          <span className={styles.clock}>
            {hrs < 10 ? "0" + hrs : hrs}:{min < 10 ? "0" + min : min}:
            {sec < 10 ? "0" + sec : sec}
          </span>
          <button
            className={timerOn ? styles.Red : styles.Blue}
            onClick={handleSubmit}
            disabled={!projectId}
          >
            {timerOn ? "Stop" : "Start"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default TaskInput;
