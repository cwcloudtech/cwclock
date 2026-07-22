import axios from "axios";
import {
  ExportJobLOADING,
  ExportJobERROR,
  ExportJobListSUCCESS,
  ExportJobCreateSUCCESS,
  ExportJobUpdateSUCCESS,
  ExportJobDeleteSUCCESS,
} from "./ExportJob.types";

export const listExportJobsApi = (orgId, token) => (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  axios
    .get(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then(({ data }) => {
      dispatch({ type: ExportJobListSUCCESS, payload: data || [] });
    })
    .catch(({ response }) => {
      dispatch({
        type: ExportJobERROR,
        payload: response?.data?.message || "Failed to load export jobs",
      });
    });
};

export const createExportJobApi = (orgId, jobData, token) => (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  axios
    .post(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs`, jobData, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then(({ data }) => {
      dispatch({ type: ExportJobCreateSUCCESS, payload: data });
    })
    .catch(({ response }) => {
      dispatch({
        type: ExportJobERROR,
        payload: response?.data?.message || "Failed to create export job",
      });
    });
};

export const updateExportJobApi = (orgId, jobId, jobData, token) => (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  axios
    .put(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs/${jobId}`, jobData, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then(({ data }) => {
      dispatch({ type: ExportJobUpdateSUCCESS, payload: data });
    })
    .catch(({ response }) => {
      dispatch({
        type: ExportJobERROR,
        payload: response?.data?.message || "Failed to update export job",
      });
    });
};

export const deleteExportJobApi = (orgId, jobId, token) => (dispatch) => {
  dispatch({ type: ExportJobLOADING });
  axios
    .delete(`${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/export-jobs/${jobId}`, {
      headers: { Authorization: `Bearer ${token}` },
    })
    .then(() => {
      dispatch({ type: ExportJobDeleteSUCCESS, payload: jobId });
    })
    .catch(({ response }) => {
      dispatch({
        type: ExportJobERROR,
        payload: response?.data?.message || "Failed to delete export job",
      });
    });
};
