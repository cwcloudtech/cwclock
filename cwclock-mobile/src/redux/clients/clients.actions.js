import client from "../../api/client";

export const ClientsSUCCESS = "clients/Success";
export const ClientsERROR = "clients/Error";

// Read-only picker data (project/invoice screens pick a client by name) -
// ported from cwclock-ui/src/Redux/Clients/Client.actions.js's listClientsApi,
// mobile has no client management screen.
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
