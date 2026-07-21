import React, { useState } from "react";
import { useDispatch } from "react-redux";
import styles from "./Styles/Form.module.css";
import { verifyMfaTotpApi, beginMfaWebAuthnLoginApi, finishMfaWebAuthnLoginApi } from "../../Redux/Auth/Auth.actions";
import { preparePublicKeyRequestOptions, assertionToJSON, webauthnSupported } from "../common/webauthnBrowser";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";

// MfaChallengeForm is the second step of login once UserHandler.Login has
// responded with an MFA challenge (see ai-instruct-68): an authenticator
// app code and/or a registered security key, depending on what the account
// has enrolled.
const MfaChallengeForm = ({ challenge }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [code, setCode] = useState("");
  const [error, setError] = useState("");
  const [verifying, setVerifying] = useState(false);

  const handleTotpSubmit = async (e) => {
    e.preventDefault();
    setError("");
    setVerifying(true);
    try {
      await dispatch(verifyMfaTotpApi(challenge.challengeToken, code));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setVerifying(false);
    }
  };

  const handleWebAuthn = async () => {
    setError("");
    setVerifying(true);
    try {
      const beginResponse = await dispatch(beginMfaWebAuthnLoginApi(challenge.challengeToken));
      const assertion = await navigator.credentials.get(preparePublicKeyRequestOptions(beginResponse.options));
      await dispatch(finishMfaWebAuthnLoginApi(beginResponse.ceremonyToken, assertionToJSON(assertion)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setVerifying(false);
    }
  };

  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{t("auth.mfaChallengeTitle")}</h1>

      {challenge.hasTotp && (
        <form onSubmit={handleTotpSubmit}>
          <div className={styles.field}>
            <input
              className={styles.input}
              type="text"
              inputMode="numeric"
              autoComplete="one-time-code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder={t("auth.mfaCodePlaceholder")}
              title={t("auth.mfaCodePlaceholder")}
              autoFocus
            />
          </div>
          <button type="submit" className={styles.btn} disabled={verifying || !code} title={t("auth.mfaVerify")}>
            {t("auth.mfaVerify")}
          </button>
        </form>
      )}

      {challenge.hasWebAuthn && webauthnSupported() && (
        <button
          type="button"
          className={styles.btn}
          style={{ marginTop: challenge.hasTotp ? 12 : 0 }}
          disabled={verifying}
          onClick={handleWebAuthn}
          title={t("auth.mfaUseSecurityKey")}
        >
          {t("auth.mfaUseSecurityKey")}
        </button>
      )}

      {error && <p className="cw-error">{error}</p>}
    </div>
  );
};

export default MfaChallengeForm;
