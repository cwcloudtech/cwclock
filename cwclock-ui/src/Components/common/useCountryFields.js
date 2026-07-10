import { useEffect, useState } from "react";
import axios from "axios";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/fields`;

// Cached per country at module scope, like useCurrencies/useCountries.
const cache = new Map();
const pending = new Map();

const fetchFields = (country) => {
  if (cache.has(country)) return Promise.resolve(cache.get(country));
  if (!pending.has(country)) {
    pending.set(
      country,
      axios
        .get(ENDPOINT, { params: { country } })
        .then(({ data }) => {
          const fields = data.fields || [];
          cache.set(country, fields);
          return fields;
        })
        .catch(() => [])
        .finally(() => {
          pending.delete(country);
        })
    );
  }
  return pending.get(country);
};

// useCountryFields fetches the business identification fields to display
// for the given country (ai-instruct-35's decision table), e.g. "FR" ->
// ["SIRET","SIREN","NAF","VAT Code"], so the create/edit forms can render
// them dynamically instead of always showing a fixed set.
const useCountryFields = (country) => {
  const [fields, setFields] = useState(cache.get(country) || []);

  useEffect(() => {
    if (!country) {
      setFields([]);
      return undefined;
    }
    let active = true;
    fetchFields(country).then((data) => {
      if (active) setFields(data);
    });
    return () => {
      active = false;
    };
  }, [country]);

  return fields;
};

export default useCountryFields;
