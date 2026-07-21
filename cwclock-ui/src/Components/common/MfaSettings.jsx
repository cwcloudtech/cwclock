import React, { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { MdDeleteForever } from "react-icons/md";
import Button from "./Button";
import Switch from "./Switch";
import Tooltip from "./Tooltip";
import ConfirmModal from "./ConfirmModal";
import SecurityKeyNameModal from "./SecurityKeyNameModal";
import {
  mfaStatusApi,
  totpSetupApi,
  totpConfirmApi,
  totpDisableApi,
  webauthnRegisterBeginApi,
  webauthnRegisterFinishApi,
  webauthnDeleteApi,
} from "../../Redux/Users/User.actions";
import { meApi } from "../../Redux/Auth/Auth.actions";
import { preparePublicKeyCreationOptions, attestationToJSON, webauthnSupported } from "./webauthnBrowser";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import styles from "./Styles/MfaSettings.module.css";

// MfaSettings is the self-service MFA enrollment section shown in
// EditProfileModal (see ai-instruct-68): enroll/disable a TOTP authenticator
// app (with QR code) and register/remove WebAuthn security keys.
const MfaSettings = ({ token }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const [status, setStatus] = useState(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [setup, setSetup] = useState(null);
  const [confirmCode, setConfirmCode] = useState("");
  const [disablingTotp, setDisablingTotp] = useState(false);
  const [deletingCredential, setDeletingCredential] = useState(null);
  const [pendingWebAuthn, setPendingWebAuthn] = useState(null);

  const refresh = async () => {
    try {
      setStatus(await dispatch(mfaStatusApi(token)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleStartTotpSetup = async () => {
    setError("");
    setBusy(true);
    try {
      setSetup(await dispatch(totpSetupApi(token)));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleConfirmTotp = async (e) => {
    e.preventDefault();
    setError("");
    setBusy(true);
    try {
      await dispatch(totpConfirmApi(confirmCode, token));
      setSetup(null);
      setConfirmCode("");
      await refresh();
      dispatch(meApi(token));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleDisableTotp = async () => {
    setError("");
    try {
      await dispatch(totpDisableApi(token));
      await refresh();
      dispatch(meApi(token));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setDisablingTotp(false);
    }
  };

  const handleRegisterWebAuthn = async () => {
    setError("");
    setBusy(true);
    try {
      const beginResponse = await dispatch(webauthnRegisterBeginApi(token));
      const credential = await navigator.credentials.create(preparePublicKeyCreationOptions(beginResponse.options));
      setPendingWebAuthn({ ceremonyToken: beginResponse.ceremonyToken, credential: attestationToJSON(credential) });
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleConfirmKeyName = async (name) => {
    setError("");
    setBusy(true);
    try {
      await dispatch(webauthnRegisterFinishApi(pendingWebAuthn.ceremonyToken, pendingWebAuthn.credential, name, token));
      setPendingWebAuthn(null);
      await refresh();
      dispatch(meApi(token));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleConfirmDeleteCredential = async () => {
    const credential = deletingCredential;
    setError("");
    try {
      await dispatch(webauthnDeleteApi(credential.id, token));
      setDeletingCredential(null);
      await refresh();
      dispatch(meApi(token));
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    }
  };

  if (!status) {
    return null;
  }

  return (
    <div className={styles.container}>
      <h3 className={styles.heading}>{t("profile.mfaTitle")}</h3>

      <div className={styles.factor}>
        <span className={styles.factorLabel}>{t("profile.mfaTotp")}</span>
        <Switch
          checked={status.totpEnabled}
          disabled={busy || (!status.totpEnabled && !!setup)}
          onChange={() => (status.totpEnabled ? setDisablingTotp(true) : handleStartTotpSetup())}
          aria-label={status.totpEnabled ? t("profile.mfaDisable") : t("profile.mfaEnable")}
          title={status.totpEnabled ? t("profile.mfaDisable") : t("profile.mfaEnable")}
        />
      </div>

      {setup && !status.totpEnabled && (
        <form onSubmit={handleConfirmTotp} className={styles.totpSetup}>
          <img src={setup.qrCodePng} alt="" className={styles.qrCode} />
          <p className={styles.secretHint}>{t("profile.mfaSecretHint", { secret: setup.secret })}</p>
          <input
            className="cw-input"
            type="text"
            inputMode="numeric"
            value={confirmCode}
            onChange={(e) => setConfirmCode(e.target.value)}
            placeholder={t("auth.mfaCodePlaceholder")}
          />
          <Button type="submit" size="sm" disabled={busy || !confirmCode}>
            {t("profile.mfaConfirm")}
          </Button>
        </form>
      )}

      {webauthnSupported() && (
        <div className={styles.factor}>
          <span className={styles.factorLabel}>{t("profile.mfaWebAuthn")}</span>
          <Button type="button" size="sm" onClick={handleRegisterWebAuthn} disabled={busy}>
            {t("profile.mfaRegisterKey")}
          </Button>
        </div>
      )}

      {status.webauthnCredentials.length > 0 && (
        <ul className={styles.keyList}>
          {status.webauthnCredentials.map((cred) => (
            <li key={cred.id} className={styles.keyRow}>
              <span className={styles.keyName}>{cred.name}</span>
              <Tooltip label={t("common.delete")}>
                <button type="button" className={styles.iconBtnDanger} onClick={() => setDeletingCredential(cred)}>
                  <MdDeleteForever style={{ fontSize: "18px" }} />
                </button>
              </Tooltip>
            </li>
          ))}
        </ul>
      )}

      {error && <p className="cw-error">{error}</p>}

      <ConfirmModal
        show={disablingTotp}
        title={t("profile.mfaDisableTitle")}
        body={t("profile.mfaDisableBody")}
        confirmLabel={t("profile.mfaDisable")}
        onConfirm={handleDisableTotp}
        onCancel={() => setDisablingTotp(false)}
      />

      <SecurityKeyNameModal
        show={!!pendingWebAuthn}
        busy={busy}
        onConfirm={handleConfirmKeyName}
        onCancel={() => setPendingWebAuthn(null)}
      />

      <ConfirmModal
        show={!!deletingCredential}
        title={t("profile.mfaDeleteKeyTitle")}
        body={deletingCredential ? t("profile.mfaDeleteKeyBody", { name: deletingCredential.name }) : ""}
        confirmLabel={t("common.delete")}
        onConfirm={handleConfirmDeleteCredential}
        onCancel={() => setDeletingCredential(null)}
      />
    </div>
  );
};

export default MfaSettings;
