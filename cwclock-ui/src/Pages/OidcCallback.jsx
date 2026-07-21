import React, { useEffect } from "react";
import axios from "axios";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import Spinner from "../Components/spinner/Spinner";
import { oidcLoginApi, oidcMfaChallengeApi } from "../Redux/Auth/Auth.actions";
import { useI18n } from "../i18n/I18nContext";
import toastOptions from "../Redux/toastOptions";

// OidcCallback is where an OIDC login (google/github/keycloak) lands once
// it's done. Two flows reach it:
//  - the API's own callback redirects here with ?token= (success), ?error=
//    (denied/failed) or ?mfaChallenge=&hasTotp=&hasWebAuthn= (the account
//    has MFA enabled - see ai-instruct-70), since it can only redirect, not
//    call the frontend origin directly;
//  - when OidcButtons asked for a frontend-bound redirect_uri, the provider
//    redirects here directly with ?code=&state=, and this page itself calls
//    the API to complete the exchange and get a token (or the same MFA
//    challenge) back as JSON.
const OidcCallback = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, isError } = useSelector((state) => state.auth);

  useEffect(() => {
    const token = searchParams.get("token");
    const oidcError = searchParams.get("error");
    const code = searchParams.get("code");
    const state = searchParams.get("state");
    const mfaChallenge = searchParams.get("mfaChallenge");

    const fail = () => {
      toast.error(t("errors.oidcFailed"), toastOptions);
      navigate("/login", { replace: true });
    };

    if (oidcError) {
      fail();
      return;
    }
    if (mfaChallenge) {
      dispatch(
        oidcMfaChallengeApi({
          challengeToken: mfaChallenge,
          hasTotp: searchParams.get("hasTotp") === "1",
          hasWebAuthn: searchParams.get("hasWebAuthn") === "1",
        })
      );
      navigate("/login", { replace: true });
      return;
    }
    if (token) {
      dispatch(oidcLoginApi(token));
      return;
    }
    if (code && state) {
      axios
        .get(`${process.env.REACT_APP_APIURL}/v1/oidc/callback`, { params: { code, state } })
        .then(({ data }) => {
          if (data?.mfaRequired) {
            dispatch(oidcMfaChallengeApi(data));
            navigate("/login", { replace: true });
            return;
          }
          if (!data?.token) {
            throw new Error("missing token");
          }
          dispatch(oidcLoginApi(data.token));
        })
        .catch(fail);
      return;
    }
    navigate("/login", { replace: true });
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
