import {
  DeleteTasksSUCCESS,
  GetTasksERROR,
  GetTasksLOADING,
  GetTasksSUCCESS,
  PostTasksSUCCESS,
  start,
  UpdateTasksSUCCESS,
} from "./Task.types";
import axios from "axios";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/time-entries/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const startTask = (payload) => (dispatch) => {
  dispatch({ type: start, payload: payload });
};
export const getTasksApi = (orgId, token) => async (dispatch) => {
  dispatch({
    type: GetTasksLOADING,
  });
  try {
    const { data } = await axios.get(ENDPOINT(orgId), authConfig(token));
    dispatch({
      type: GetTasksSUCCESS,
      payload: data,
    });
    return data;
  } catch (e) {
    dispatch({
      type: GetTasksERROR,
    });
  }
};
// Time record creation intentionally has no success toast: it fires every
// time the timer is stopped, which would be too noisy.
export const postTasksApi = (item, orgId, token) => async (dispatch) => {
  dispatch({ type: GetTasksLOADING });
  try {
    const { data } = await axios.post(ENDPOINT(orgId), item, authConfig(token));
    dispatch({ type: PostTasksSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: GetTasksERROR });
  }
};
export const deleteTasksApi = (id, orgId, token) => async (dispatch) => {
  dispatch({
    type: GetTasksLOADING,
  });
  try {
    await axios.delete(ENDPOINT(orgId) + id, authConfig(token));
    dispatch({ type: DeleteTasksSUCCESS, payload: id });
    toast.success("Time record deleted.", toastOptions);
  } catch (e) {
    dispatch({
      type: GetTasksERROR,
    });
  }
};
export const updateTasksApi = (task, orgId, token) => async (dispatch) => {
  dispatch({
    type: GetTasksLOADING,
  });
  try {
    const { id } = task;
    const { data } = await axios.put(ENDPOINT(orgId) + id, task, authConfig(token));
    dispatch({ type: UpdateTasksSUCCESS, payload: data });
    toast.success("Time record updated.", toastOptions);
  } catch (e) {
    dispatch({
      type: GetTasksERROR,
    });
  }
};
