import { useEffect, useState } from "react";
import axios from "axios";

// useAppVersion fetches manifest.json - copied next to the built frontend at
// deploy time (see Dockerfile), not served by the API, so it's fetched from
// the app's own origin instead of REACT_APP_APIURL. Shared by Slidebar.jsx's
// version badge/Android download link and AuthNav.jsx's Android download
// link (login/signup forms, ai-instruct-106).
const useAppVersion = () => {
  const [appVersion, setAppVersion] = useState(null);

  useEffect(() => {
    axios
      .get("/manifest.json")
      .then(({ data }) => setAppVersion(data.version))
      .catch(() => {});
  }, []);

  return appVersion;
};

export default useAppVersion;
