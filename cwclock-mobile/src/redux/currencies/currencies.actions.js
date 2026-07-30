import client from "../../api/client";

export const CurrenciesSUCCESS = "currencies/Success";

// GET /v1/currencies is public (no auth, no org scope), returns {
// currencies: [{ iso, name }] } - see countries.actions.js's listCountriesApi
// for the equivalent country list and why this is a plain Redux action here
// rather than the web app's cached hook.
export const listCurrenciesApi = () => async (dispatch) => {
  const { data } = await client.get("/currencies");
  dispatch({ type: CurrenciesSUCCESS, payload: data?.currencies });
};
