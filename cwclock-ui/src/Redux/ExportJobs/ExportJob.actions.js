import axios from "axios";
import { toast } from "react-toastify";
import {
  ExportJobLOADING,
  ExportJobERROR,
  ExportJobListSUCCESS,
  ExportJobCreateSUCCESS,
  ExportJobUpdateSUCCESS,
  ExportJobDeleteSUCCESS,
} from "./ExportJob.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

export const listExportJobsApi = (orgId, token) => async (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  try {
    const { data } = await axios.get(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: ExportJobListSUCCESS, payload: data || [] });
  } catch (e) {
    dispatch({ type: ExportJobERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
  }
};

export const createExportJobApi = (orgId, jobData, token) => async (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  try {
    const { data } = await axios.post(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs`, jobData, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: ExportJobCreateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "exportJobs.createdSuccessfully"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ExportJobERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const updateExportJobApi = (orgId, jobId, jobData, token) => async (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  try {
    const { data } = await axios.put(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs/${jobId}`, jobData, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: ExportJobUpdateSUCCESS, payload: data });
    toast.success(translate(getStoredLocale(), "exportJobs.updatedSuccessfully"), toastOptions);
    return data;
  } catch (e) {
    dispatch({ type: ExportJobERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

export const deleteExportJobApi = (orgId, jobId, token) => async (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  try {
    await axios.delete(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs/${jobId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: ExportJobDeleteSUCCESS, payload: jobId });
    toast.success(translate(getStoredLocale(), "exportJobs.deletedSuccessfully"), toastOptions);
  } catch (e) {
    dispatch({ type: ExportJobERROR });
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
