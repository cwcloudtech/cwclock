import client from "../../api/client";

export const ProjectsSUCCESS = "projects/Success";
export const ProjectsERROR = "projects/Error";
export const ProjectCreateSUCCESS = "projects/CreateSuccess";
export const ProjectUpdateSUCCESS = "projects/UpdateSuccess";
export const ProjectDeleteSUCCESS = "projects/DeleteSuccess";

// Ported from cwclock-ui/src/Redux/Projects/Project.actions.js - read-only
// picker data (project's clientId is what TimeTrackerScreen/
// EditRecordScreen/AllDayRecordScreen infer a time entry's client from,
// matching the web app's TaskInput.jsx behavior) plus, since ai-instruct-106,
// full CRUD for ProjectsScreen/ProjectFormScreen.
export const listProjectsApi = (orgId) => async (dispatch) => {
  try {
    const { data } = await client.get(`/organizations/${orgId}/projects/`);
    dispatch({ type: ProjectsSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ProjectsERROR });
    throw e;
  }
};

// createProjectApi creates a project under a specific client - the API
// nests project creation under its client (POST .../clients/{clientId}/
// projects/), unlike update/delete which address the project directly.
export const createProjectApi = (orgId, clientId, fields) => async (dispatch) => {
  const { data } = await client.post(`/organizations/${orgId}/clients/${clientId}/projects/`, fields);
  dispatch({ type: ProjectCreateSUCCESS, payload: data });
  return data;
};

// updateProjectApi's fields includes clientId - the API allows moving a
// project to a different client of the same organization this way (see
// ProjectHandler.Update), not just editing its own fields.
export const updateProjectApi = (orgId, projectId, fields) => async (dispatch) => {
  const { data } = await client.put(`/organizations/${orgId}/projects/${projectId}`, fields);
  dispatch({ type: ProjectUpdateSUCCESS, payload: data });
  return data;
};

export const deleteProjectApi = (orgId, projectId) => async (dispatch) => {
  await client.delete(`/organizations/${orgId}/projects/${projectId}`);
  dispatch({ type: ProjectDeleteSUCCESS, payload: projectId });
};
