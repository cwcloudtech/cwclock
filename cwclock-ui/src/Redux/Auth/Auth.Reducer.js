import { error, loading, login, logout, register, updatePicture, updateProfile, syncUser, mfaRequired } from "./Auth.types";

const user = JSON.parse(localStorage.getItem("User"));

const initialstate = {
  user: user || {},
  isError: false,
  isLoading: false,
  message: "",
  mfaChallenge: null,
};

export const AuthReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case loading: {
      return { ...state, isLoading: true };
    }
    case error: {
      return { ...state, isLoading: false, isError: true };
    }
    case register: {
      localStorage.setItem("User", JSON.stringify(payload));
      return {
        ...state,
        isLoading: false,
        user: { ...payload },
        isError: false,
        mfaChallenge: null,
      };
    }
    case login: {
      localStorage.setItem("User", JSON.stringify(payload));
      return {
        ...state,
        isLoading: false,
        user: { ...payload },
        isError: false,
        mfaChallenge: null,
      };
    }
    case logout: {
      localStorage.removeItem("User");
      return { ...state, isLoading: false, user: {}, isError: false, mfaChallenge: null };
    }
    // mfaRequired intentionally never touches localStorage/user - the
    // password was correct but login isn't complete until the second
    // factor is verified (see Auth.types.js).
    case mfaRequired: {
      return { ...state, isLoading: false, isError: false, mfaChallenge: payload };
    }
    case updatePicture: {
      const updatedUser = { ...state.user, picture: payload.picture };
      localStorage.setItem("User", JSON.stringify(updatedUser));
      return { ...state, user: updatedUser };
    }
    case updateProfile: {
      const updatedUser = { ...state.user, name: payload.name, surname: payload.surname };
      localStorage.setItem("User", JSON.stringify(updatedUser));
      return { ...state, user: updatedUser };
    }
    case syncUser: {
      // Merge the latest /me snapshot (role can change server-side, eg. the
      // superuser confirms or disables the account) without a token field,
      // which /me never returns, so the existing token is preserved.
      const updatedUser = { ...state.user, ...payload };
      localStorage.setItem("User", JSON.stringify(updatedUser));
      return { ...state, user: updatedUser };
    }
    default: {
      return state;
    }
  }
};
