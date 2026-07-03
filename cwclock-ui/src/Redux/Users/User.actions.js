import axios from "axios";
import { updatePicture } from "../Auth/Auth.types";

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
