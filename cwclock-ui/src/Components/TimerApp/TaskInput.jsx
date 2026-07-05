import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import styles from "./Styles/TaskInput.module.css";
import useTimer from "./useTimer";
import useDateHook from "./useDateHook";
import { postTasksApi, startTask } from "../../Redux/Tasks/Task.actions";
import useTime from "./useTime";
import projectLabel from "../common/projectLabel";

const TaskInput = () => {
  const { timerOn, sec, min, hrs, handleTimer } = useTimer();
  const { hours, minutes, seconds } = useDateHook();
  const { hours2, minutes2, seconds2 } = useTime();
  const [name, setName] = useState("");
  const [projectId, setProjectId] = useState("");
  const [projectQuery, setProjectQuery] = useState("");
  const { start } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const { clients } = useSelector((state) => state.clients);
  const dispatch = useDispatch();

  const handleProjectInput = (text) => {
    setProjectQuery(text);
    const match = projects.find((p) => projectLabel(p, clients) === text);
    setProjectId(match ? match.id : "");
  };

  const handleSubmit = () => {
    if (timerOn) {
      const project = projects.find((p) => p.id === projectId);
      if (!project) {
        return;
      }
      let taskObj = {
        text: name || "Task",
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
          title="What are you working on?"
          onChange={(e) => setName(e.target.value)}
        />
        <input
          list="task-input-project-options"
          className={styles.Projects}
          value={projectQuery}
          disabled={timerOn}
          onChange={(e) => handleProjectInput(e.target.value)}
          placeholder="Project"
          title="Search by customer or project name"
        />
        <datalist id="task-input-project-options">
          {projects.map((p) => (
            <option key={p.id} value={projectLabel(p, clients)} />
          ))}
        </datalist>
        <div className={styles.Timer}>
          <span className={styles.clock} title="Elapsed time">
            {hrs < 10 ? "0" + hrs : hrs}:{min < 10 ? "0" + min : min}:
            {sec < 10 ? "0" + sec : sec}
          </span>
          <button
            className={timerOn ? styles.Red : styles.Blue}
            onClick={handleSubmit}
            disabled={!projectId}
            title={timerOn ? "Stop the timer" : "Start the timer"}
          >
            {timerOn ? "Stop" : "Start"}
          </button>
        </div>
      </div>
    </div>
  );
};

export default TaskInput;
