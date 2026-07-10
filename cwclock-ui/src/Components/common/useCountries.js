import { useEffect, useState } from "react";
import axios from "axios";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/countries`;

// Cached at module scope so every component sharing this hook fires a single
// request instead of one per mount.
let cache = null;
let pending = null;

const fetchCountries = () => {
  if (cache) return Promise.resolve(cache);
  if (!pending) {
    pending = axios
      .get(ENDPOINT)
      .then(({ data }) => {
        cache = data.countries || [];
        return cache;
      })
      .catch(() => [])
      .finally(() => {
        pending = null;
      });
  }
  return pending;
};

// useCountries fetches the countries table's list of { iso, name, currency }
// (see ai-instruct-35), used to power the country autocomplete on
// organization/client forms instead of a free-text input.
const useCountries = () => {
  const [countries, setCountries] = useState(cache || []);

  useEffect(() => {
    let active = true;
    fetchCountries().then((data) => {
      if (active) setCountries(data);
    });
    return () => {
      active = false;
    };
  }, []);

  return countries;
};

export default useCountries;
