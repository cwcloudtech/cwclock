import axios from "axios";
import { updatePicture, updateProfile } from "../Auth/Auth.types";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/users/`;

export const updatePictureApi = (picture, token) => async (dispatch) => {
  try {
    const { data } = await axios.put(
      `${ENDPOINT}me/picture`,
      { picture },
      { headers: { Authorization: `Bearer ${token}` } }
    );
    dispatch({ type: updatePicture, payload: data });
    return data;
  } catch (e) {
    // Swallow: the profile dropdown just keeps the previous picture.
  }
};

export const updateProfileApi = (name, surname, token) => async (dispatch) => {
  const { data } = await axios.put(
    `${ENDPOINT}me`,
    { name, surname },
    { headers: { Authorization: `Bearer ${token}` } }
  );
  dispatch({ type: updateProfile, payload: data });
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
