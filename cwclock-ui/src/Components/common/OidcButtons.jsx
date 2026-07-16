import React, { useEffect, useState } from "react";
import axios from "axios";
import { toast } from "react-toastify";
import { FaGoogle, FaGithub, FaKey } from "react-icons/fa";
import { useI18n } from "../../i18n/I18nContext";
import toastOptions from "../../Redux/toastOptions";
import styles from "./Styles/OidcButtons.module.css";

const ENDPOINT = `${process.env.REACT_APP_APIURL}/v1/oidc`;

// PROVIDERS maps each backend provider id to the icon/label/brand color used
// to render its button; order here is the display order.
const PROVIDERS = [
  { id: "google", icon: FaGoogle, labelKey: "auth.continueWithGoogle", className: "google" },
  { id: "github", icon: FaGithub, labelKey: "auth.continueWithGithub", className: "github" },
  { id: "keycloak", icon: FaKey, labelKey: "auth.continueWithKeycloak", className: "keycloak" },
];

// OidcButtons shows one login/signup button per OIDC provider the backend
// has configured (GET /v1/oidc), each linking to its server-side redirect
// flow. Renders nothing while loading or when no provider is configured, so
// deployments without any CWCLOCK_OIDC_* env vars set see no visual change.
const OidcButtons = () => {
  const { t } = useI18n();
  const [providers, setProviders] = useState([]);

  useEffect(() => {
    let cancelled = false;
    axios
      .get(ENDPOINT)
      .then(({ data }) => {
        if (!cancelled) setProviders(data?.providers || []);
      })
      .catch(() => {
        if (!cancelled) setProviders([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  if (providers.length === 0) {
    return null;
  }

  // Asks the API for a frontend-bound authorization URL (X-CWClock-Origin
  // tells it to point redirect_uri at our own /oidc/callback instead of its
  // own) and navigates the browser there ourselves; a plain <a href> can't
  // attach that header, so the link is a progressive-enhancement fallback
  // for JS-less/middle-click navigation (still works via the API's default
  // redirect_uri flow).
  const handleLogin = (id) => async (e) => {
    e.preventDefault();
    try {
      const { data } = await axios.get(`${process.env.REACT_APP_APIURL}/v1/oidc/${id}/login`, {
        headers: { "X-CWClock-Origin": "frontend" },
      });
      if (data?.url) {
        window.location.href = data.url;
      }
    } catch (err) {
      toast.error(t("errors.oidcFailed"), toastOptions);
    }
  };

  return (
    <div className={styles.wrapper}>
      <div className={styles.divider}>
        <span>{t("auth.orContinueWith")}</span>
      </div>
      <div className={styles.buttons}>
        {PROVIDERS.filter((p) => providers.includes(p.id)).map(({ id, icon: Icon, labelKey, className }) => (
          <a
            key={id}
            href={`${process.env.REACT_APP_APIURL}/v1/oidc/${id}/login`}
            onClick={handleLogin(id)}
            className={`${styles.btn} ${styles[className]}`}
            title={t(labelKey)}
          >
            <Icon className={styles.icon} />
            <span>{t(labelKey)}</span>
          </a>
        ))}
      </div>
    </div>
  );
};

export default OidcButtons;
