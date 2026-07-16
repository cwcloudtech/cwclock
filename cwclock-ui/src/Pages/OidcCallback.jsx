import React, { useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import Spinner from "../Components/spinner/Spinner";
import { oidcLoginApi } from "../Redux/Auth/Auth.actions";
import { useI18n } from "../i18n/I18nContext";
import toastOptions from "../Redux/toastOptions";

// OidcCallback is where the backend redirects the browser once an OIDC
// login (google/github/keycloak) completes: it carries either ?token= (on
// success) or ?error= (denied/failed) in the query string, since the
// backend's own callback endpoint can only redirect, not call the frontend
// origin directly.
const OidcCallback = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, isError } = useSelector((state) => state.auth);

  useEffect(() => {
    const token = searchParams.get("token");
    const oidcError = searchParams.get("error");

    if (oidcError) {
      toast.error(t("errors.oidcFailed"), toastOptions);
      navigate("/login", { replace: true });
      return;
    }
    if (token) {
      dispatch(oidcLoginApi(token));
    } else {
      navigate("/login", { replace: true });
    }
    // Runs once: the query string is only ever meaningful on the initial load of this page.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (user.token) {
      navigate("/dashboard/timetracker", { replace: true });
    } else if (isError) {
      navigate("/login", { replace: true });
    }
  }, [user, isError, navigate]);

  return <Spinner />;
};

export default OidcCallback;
