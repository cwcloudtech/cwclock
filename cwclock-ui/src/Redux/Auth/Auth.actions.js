import axios from "axios";
import { error, loading, logout, register } from "./Auth.types";
import { toast } from "react-toastify";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/`;

const toastOptions = {
  autoClose: 8000,
  draggable: true,
};

export const registerApi = (userData) => async (dispatch) => {
  dispatch({ type: loading });
  try {
    console.log(userData);
    const { data } = await axios.post(ENDPOINT, userData);
    dispatch({ type: register, payload: data });
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
  } catch (e) {
    toast.error(e.message, toastOptions);
    dispatch({ type: error, payload: e.message });
  }
};

export const logoutUser = () => (dispatch) => {
  dispatch({ type: logout });
};

// meApi verifies that the connected user still exists in the database,
// disconnecting them otherwise (eg. their account was deleted elsewhere).
export const meApi = (token) => async (dispatch) => {
  try {
    await axios.get(`${ENDPOINT}me`, {
      headers: { Authorization: `Bearer ${token}` },
    });
  } catch (e) {
    dispatch({ type: logout });
  }
};
