import axios from "axios";
import { toast } from "react-toastify";
import { ProjectLOADING, ProjectERROR, ProjectListSUCCESS, ProjectCreateSUCCESS } from "./Project.types";
import toastOptions from "../toastOptions";

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

export const createProjectApi = (orgId, clientId, name, color, token) => async (dispatch) => {
  dispatch({ type: ProjectLOADING });
  try {
    const { data } = await axios.post(
      `${ENDPOINT}${orgId}/clients/${clientId}/projects/`,
      { name, color },
      authConfig(token)
    );
    dispatch({ type: ProjectCreateSUCCESS, payload: data });
    toast.success("Project created.", toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ProjectERROR });
    throw e;
  }
};
