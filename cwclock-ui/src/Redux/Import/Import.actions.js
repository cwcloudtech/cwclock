import axios from "axios";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";
import { apiErrorMessage, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = (orgId) => `${process.env.REACT_APP_APIURL}/v1/organizations/${orgId}/import/csv`;

// No dedicated reducer: the import result (created/skipped counts) is only
// ever shown once, right after the request, so the calling component keeps
// it in local state instead of global store state.
export const importCSVApi = (orgId, csvText, token) => async () => {
  try {
    const { data } = await axios.post(ENDPOINT(orgId), csvText, {
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "text/csv" },
    });
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
