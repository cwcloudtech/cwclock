import React from "react";
import TaskInput from "./TaskInput";
import { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import TaskComponent from "./TaskComponent";
import Heading from "./Heading";
import styles from "./Styles/TaskApp.module.css";
import EmptyTask from "./EmptyTask";
import { useNavigate, Link } from "react-router-dom";
import Spinner from "../spinner/Spinner";
import { getTasksApi } from "../../Redux/Tasks/Task.actions";
import { listOrgsApi, listMembersApi } from "../../Redux/Organizations/Org.actions";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";

const TasksApp = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { tasks, isLoading } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);

  useEffect(() => {
    if (user.token) {
      dispatch(listOrgsApi(user.token));
    } else {
      navigate("/login");
    }
  }, [user, dispatch, navigate]);

  useEffect(() => {
    if (user.token && currentOrgId) {
      dispatch(getTasksApi(currentOrgId, user.token));
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [user, currentOrgId, dispatch]);

  if (isLoading) {
    return (
      <div className={styles.Body1}>
        <h1 className={styles.Load1}>
          <Spinner />
        </h1>
      </div>
    );
  }

  if (!currentOrgId) {
    return (
      <div className={styles.Body1}>
        <p>
          You need an organization before you can track time.{" "}
          <Link to="/dashboard/organizations">Create one</Link>.
        </p>
      </div>
    );
  }

  return (
    <div className={styles.Body1}>
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
    </div>
  );
};

export default TasksApp;
