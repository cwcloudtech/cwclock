import client from "../../api/client";

export const ClientsSUCCESS = "clients/Success";
export const ClientsERROR = "clients/Error";
export const ClientCreateSUCCESS = "clients/CreateSuccess";
export const ClientUpdateSUCCESS = "clients/UpdateSuccess";
export const ClientDeleteSUCCESS = "clients/DeleteSuccess";

// Ported from cwclock-ui/src/Redux/Clients/Client.actions.js - same
// endpoints/payload shape as the web app's client management (Clients.jsx/
// EditClientModal.jsx), minus the country-specific identification fields
// (SIREN/SIRET/NAF/MF - see useCountryFields.js) and cross-organization
// transfer, which mobile's forms don't expose (ai-instruct-106 scope).
export const listClientsApi = (orgId) => async (dispatch) => {
  try {
    const { data } = await client.get(`/organizations/${orgId}/clients/`);
    dispatch({ type: ClientsSUCCESS, payload: data });
    return data;
  } catch (e) {
    dispatch({ type: ClientsERROR });
    throw e;
  }
};

export const createClientApi = (orgId, fields) => async (dispatch) => {
  const { data } = await client.post(`/organizations/${orgId}/clients/`, fields);
  dispatch({ type: ClientCreateSUCCESS, payload: data });
  return data;
};

export const updateClientApi = (orgId, clientId, fields) => async (dispatch) => {
  const { data } = await client.put(`/organizations/${orgId}/clients/${clientId}`, fields);
  dispatch({ type: ClientUpdateSUCCESS, payload: data });
  return data;
};

export const deleteClientApi = (orgId, clientId) => async (dispatch) => {
  await client.delete(`/organizations/${orgId}/clients/${clientId}`);
  dispatch({ type: ClientDeleteSUCCESS, payload: clientId });
};
