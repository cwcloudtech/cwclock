import axios from "axios";
import { error, loading, logout, register, syncUser, mfaRequired } from "./Auth.types";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";
import { translate, getStoredLocale, apiErrorMessage } from "../../i18n/translate";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/`;

export const registerApi = (userData) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    console.log(userData);
    const { data } = await axios.post(ENDPOINT, userData);
    dispatch({ type: register, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.accountCreated"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

export const loginApi = (userData) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    const { data } = await axios.post(ENDPOINT + "login", userData);
    if (data.mfaRequired) {
      // Password was correct, but a second factor is required (see
      // ai-instruct-68) - nothing is persisted yet, LoginForm renders the
      // MFA challenge step next.
      dispatch({ type: mfaRequired, payload: data });
      return;
    }
    dispatch({ type: register, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.loggedIn"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

// verifyMfaTotpApi finishes a password login gated by MFA using an
// authenticator app code.
export const verifyMfaTotpApi = (challengeToken, code) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    const { data } = await axios.post(`${ENDPOINT}login/mfa/totp`, { challengeToken, code });
    dispatch({ type: register, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.loggedIn"), toastOptions);
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    dispatch({ type: error, payload: e.message });
    throw e;
  }
};

// beginMfaWebAuthnLoginApi requests the assertion options for finishing a
// password login gated by MFA using a registered security key.
export const beginMfaWebAuthnLoginApi = (challengeToken) => async () => {
  const { data } = await axios.post(`${ENDPOINT}login/mfa/webauthn/begin`, { challengeToken });
  return data;
};

// finishMfaWebAuthnLoginApi submits the security key's signed assertion to
// complete the login.
export const finishMfaWebAuthnLoginApi = (ceremonyToken, credential) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    const { data } = await axios.post(`${ENDPOINT}login/mfa/webauthn/finish`, { ceremonyToken, credential });
    dispatch({ type: register, payload: data });
    toast.success(translate(getStoredLocale(), "toasts.loggedIn"), toastOptions);
    return data;
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    dispatch({ type: error, payload: e.message });
    throw e;
  }
};

export const logoutUser = () => (dispatch) => {
  dispatch({ type: logout });
};

// oidcLoginApi finishes an OIDC login: the backend redirect only carries a
// bare session token (it never talks to the frontend origin directly), so
// the full profile is fetched from /me the same way meApi refreshes it, then
// stored exactly like a password login/register response.
export const oidcLoginApi = (token) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    const { data } = await axios.get(`${ENDPOINT}me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: register, payload: { ...data, token } });
    toast.success(translate(getStoredLocale(), "toasts.loggedIn"), toastOptions);
  } catch (e) {
    toast.error(translate(getStoredLocale(), "errors.oidcFailed"), toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

// forgotPasswordApi requests a password-renewal email for the given
// address. The backend always responds success regardless of whether the
// address is registered, so this never reveals which emails have accounts.
export const forgotPasswordApi = (email) => async () => {
  try {
    await axios.post(`${ENDPOINT}forgot-password`, { email });
    toast.success(translate(getStoredLocale(), "toasts.passwordResetSent"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// resetPasswordApi sets a new password from the token emailed by
// forgotPasswordApi.
export const resetPasswordApi = (token, password, confirmPassword) => async () => {
  try {
    await axios.post(`${ENDPOINT}reset-password`, { token, password, confirmPassword });
    toast.success(translate(getStoredLocale(), "toasts.passwordReset"), toastOptions);
  } catch (e) {
    toast.error(apiErrorMessage(e, getStoredLocale()), toastOptions);
    throw e;
  }
};

// meApi verifies that the connected user still exists in the database, and
// refreshes their locally cached profile (in particular the global role,
// which can change server-side after the token was issued), disconnecting
// them if the account no longer exists.
export const meApi = (token) => async (dispatch) => {
  try {
    const { data } = await axios.get(`${ENDPOINT}me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    dispatch({ type: syncUser, payload: data });
  } catch (e) {
    dispatch({ type: logout });
  }
};
