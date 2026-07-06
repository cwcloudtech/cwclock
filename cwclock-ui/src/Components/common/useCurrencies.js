import { useEffect, useState } from "react";
import axios from "axios";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/currencies`;

// Cached at module scope so every component sharing this hook fires a single
// request instead of one per mount.
let cache = null;
let pending = null;

const fetchCurrencies = () => {
  if (cache) return Promise.resolve(cache);
  if (!pending) {
    pending = axios
      .get(ENDPOINT)
      .then(({ data }) => {
        cache = data;
        return data;
      })
      .catch(() => [])
      .finally(() => {
        pending = null;
      });
  }
  return pending;
};

// useCurrencies fetches the backend's allowed-currency list (server-side
// default, or the CWCLOCK_ALLOWED_CURRENCIES override) instead of hardcoding
// it in the frontend.
const useCurrencies = () => {
  const [currencies, setCurrencies] = useState(cache || []);

  useEffect(() => {
    let active = true;
    fetchCurrencies().then((data) => {
      if (active) setCurrencies(data);
    });
    return () => {
      active = false;
    };
  }, []);

  return currencies;
};

export default useCurrencies;
