import axios from "axios";
import { error, loading, logout, register, syncUser } from "./Auth.types";
import { toast } from "react-toastify";
import toastOptions from "../toastOptions";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/`;

export const registerApi = (userData) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    console.log(userData);
    const { data } = await axios.post(ENDPOINT, userData);
    dispatch({ type: register, payload: data });
    toast.success("Account created successfully.", toastOptions);
  } catch (e) {
    toast.error(e.message, toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

export const loginApi = (userData) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    const { data } = await axios.post(ENDPOINT + "login", userData);
    console.log(data);
    dispatch({ type: register, payload: data });
    toast.success("Logged in successfully.", toastOptions);
  } catch (e) {
    toast.error(e.message, toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

export const logoutUser = () => (dispatch) => {
  dispatch({ type: logout });
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
