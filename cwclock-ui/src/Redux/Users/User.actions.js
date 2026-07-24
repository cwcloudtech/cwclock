import axios from "axios";
import { toast } from "react-toastify";
import { updatePicture, updateProfile } from "../Auth/Auth.types";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/`;

export const updatePictureApi = (picture, x, y, token) => async (dispatch) => {
  try {
    const { data } = await axios.put(
      `${ENDPOINT}me/picture`,
      { picture, x, y },
      { headers: { Authorization: `Bearer ${token}` } }
    );
    dispatch({ type: updatePicture, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.profilePictureUpdated"), toastOptions);
    return data;
  } catch (e) {
    // Swallow: the profile dropdown just keeps the previous picture.
  }
};

export const updateProfileApi = (name, surname, password, confirmPassword, token) => async (dispatch) => {
  const { data } = await axios.put(
    `${ENDPOINT}me`,
    { name, surname, password, confirmPassword },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  dispatch({ type: updateProfile, payload: data });
  toast.success(translate(getStoredLocale(), "toasts.profileUpdated"), toastOptions);
  return data;
};

// searchUsersApi powers email autocomplete; it doesn't touch global state,
// it just resolves with the matching users.
export const searchUsersApi = (query, token) => async () => {
  try {
    const { data } = await axios.get(`${ENDPOINT}search`, {
      params: { q: query },
      headers: { Authorization: `Bearer ${token}` },
    });
    return data;
  } catch (e) {
    return [];
  }
};

// MFA self-service enrollment (see ai-instruct-68). None of these touch
// global auth state directly - callers refetch /me (meApi) afterwards to
// refresh user.mfaEnabled, the same way other profile edits do.
const authConfig = (token) => ({ headers: { Authorization: `Bearer ${token}` } });

export const mfaStatusApi = (token) => async () => {
  const { data } = await axios.get(`${ENDPOINT}me/mfa`, authConfig(token));
  return data;
};

export const totpSetupApi = (token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/mfa/totp/setup`, {}, authConfig(token));
  return data;
};

export const totpConfirmApi = (code, token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/mfa/totp/confirm`, { code }, authConfig(token));
  return data;
};

export const totpDisableApi = (token) => async () => {
  const { data } = await axios.delete(`${ENDPOINT}me/mfa/totp`, authConfig(token));
  return data;
};

export const webauthnRegisterBeginApi = (token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/mfa/webauthn/register/begin`, {}, authConfig(token));
  return data;
};

export const webauthnRegisterFinishApi = (ceremonyToken, credential, name, token) => async () => {
  const { data } = await axios.post(
    `${ENDPOINT}me/mfa/webauthn/register/finish`,
    { ceremonyToken, credential, name },
    authConfig(token)
  );
  return data;
};

export const webauthnDeleteApi = (credentialId, token) => async () => {
  await axios.delete(`${ENDPOINT}me/mfa/webauthn/${credentialId}`, authConfig(token));
};

// Calendar-sharing feed self-service (see ai-instruct-85): an ICS URL the
// user can subscribe to from Outlook/Google Calendar. Like the MFA calls
// above, these don't touch global state - the Calendar view's share modal
// keeps the {enabled, url} status in its own local state.
export const calendarFeedStatusApi = (token) => async () => {
  const { data } = await axios.get(`${ENDPOINT}me/calendar-feed`, authConfig(token));
  return data;
};

export const calendarFeedEnableApi = (token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/calendar-feed/enable`, {}, authConfig(token));
  return data;
};

export const calendarFeedDisableApi = (token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/calendar-feed/disable`, {}, authConfig(token));
  return data;
};

export const calendarFeedRegenerateApi = (token) => async () => {
  const { data } = await axios.post(`${ENDPOINT}me/calendar-feed/regenerate`, {}, authConfig(token));
  return data;
};
