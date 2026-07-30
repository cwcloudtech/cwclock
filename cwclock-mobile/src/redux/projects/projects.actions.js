import client from "../../api/client";

export const ProjectsSUCCESS = "projects/Success";
export const ProjectsERROR = "projects/Error";

// Ported from cwclock-ui/src/Redux/Projects/Project.actions.js's
// listProjectsApi - read-only picker data (project's clientId is what
// TimeTrackerScreen/EditRecordScreen/AllDayRecordScreen infer a time
// entry's client from, matching the web app's TaskInput.jsx behavior).
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
