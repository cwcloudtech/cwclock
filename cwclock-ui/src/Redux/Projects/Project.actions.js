import axios from "axios";
import { toast } from "react-toastify";
import {
  ProjectLOADING,
  ProjectERROR,
  ProjectListSUCCESS,
  ProjectCreateSUCCESS,
  ProjectUpdateSUCCESS,
  ProjectDeleteSUCCESS,
} from "./Project.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/organizations/`;

const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const listProjectsApi = (orgId, token) => async (dispatch) => {
  dispatch({ type: ProjectLOADING });
  try {
    const { data } = await axios.get(`${ENDPOINT}${orgId}/projects/`, authConfig(token));
    dispatch({ type: ProjectListSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ProjectERROR });
  }
};

export const createProjectApi = (orgId, clientId, name, color, dailyRate, subdivisions, token) => async (dispatch) => {
  dispatch({ type: ProjectLOADING });
  try {
    const { data } = await axios.post(
      `${ENDPOINT}${orgId}/clients/${clientId}/projects/`,
      { name, color, dailyRate, subdivisions },
      authConfig(token)
    );
    dispatch({ type: ProjectCreateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.projectCreated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ProjectERROR });
    throw e;
  }
};

export const updateProjectApi = (orgId, projectId, name, color, dailyRate, subdivisions, token) => async (dispatch) => {
  dispatch({ type: ProjectLOADING });
  try {
    const { data } = await axios.put(
      `${ENDPOINT}${orgId}/projects/${projectId}`,
      { name, color, dailyRate, subdivisions },
      authConfig(token)
    );
    dispatch({ type: ProjectUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.projectUpdated"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ProjectERROR });
    throw e;
  }
};

export const deleteProjectApi = (orgId, projectId, token) => async (dispatch) => {
  dispatch({ type: ProjectLOADING });
  try {
    await axios.delete(`${ENDPOINT}${orgId}/projects/${projectId}`, authConfig(token));
    dispatch({ type: ProjectDeleteSUCCESS, payload: projectId });
    toast.success(translate(getStoredLocale(), "toasts.projectDeleted"), toastOptions);
  } catch (e) {
    dispatch({ type: ProjectERROR });
    throw e;
  }
};
