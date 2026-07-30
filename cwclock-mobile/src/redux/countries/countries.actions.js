import client from "../../api/client";

export const CountriesSUCCESS = "countries/Success";

// GET /v1/countries is public (no auth, no org scope - see
// cwclock-api/internal/router/router.go), returns { countries: [{ iso, name,
// currency }] } - same list cwclock-ui's useCountries hook powers its
// country autocomplete with (ai-instruct-35). Fetched once per screen that
// needs it (organization/client forms) rather than cached at module scope
// like the web hook, to keep this a plain Redux action like everything else
// in cwclock-mobile.
export const listCountriesApi = () => async (dispatch) => {
  const { data } = await client.get("/countries");
  dispatch({ type: CountriesSUCCESS, payload: data?.countries });
};
