import axios from "axios";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/contact`;

// sendContactApi submits the public contact form. fields: { email, subject,
// message, name, firstname } - name/firstname are optional.
export const sendContactApi = (fields) => async () => {
  try {
    await axios.post(ENDPOINT, fields);
    toast.success(translate(getStoredLocale(), "toasts.contactSent"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};
