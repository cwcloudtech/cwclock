import { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { searchUsersApi } from "../../Redux/Users/User.actions";

// Debounced email search used to power invite/transfer-ownership autocompletes.
const useEmailAutocomplete = (email, enabled, token) => {
  const dispatch = useDispatch();
  const [suggestions, setSuggestions] = useState([]);

  useEffect(() => {
    if (!enabled || email.length < 2) {
      setSuggestions([]);
      return;
    }
    const timeout = setTimeout(async () => {
      const results = await dispatch(searchUsersApi(email, token));
      setSuggestions(results || []);
    }, 300);
    return () => clearTimeout(timeout);
  }, [email, enabled, token, dispatch]);

  return suggestions;
};

export default useEmailAutocomplete;
