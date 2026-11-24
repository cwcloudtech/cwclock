import React, { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import Modal from "../common/Modal";
import AutocompleteSelect from "../common/AutocompleteSelect";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import toastOptions from "../../Redux/toastOptions";
import { workingDaysApi, fillWorkingDaysApi, getTasksApi } from "../../Redux/Tasks/Task.actions";
import styles from "./Styles/FillWorkingDaysModal.module.css";

// Format a Date as a local "YYYY-MM-DD" day key. NOT toISOString(), which
// converts to UTC and in a positive-offset timezone (e.g. France) shifts a
// local-midnight date back a day — that would push the month boundaries off by
// one (the 1st into the previous month, the last day one short).
const fmtDay = (d) => {
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${d.getFullYear()}-${month}-${day}`;
};

// Fill every working day in a date range with an all-day entry for a client /
// project, skipping weekends, the resolved country's public holidays, and days
// already logged on that project. A read-only preview (working-days endpoint)
// shows exactly what will be created; nothing is written until "Create".
const FillWorkingDaysModal = ({ show, onClose }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);

  const now = new Date();
  const [clientId, setClientId] = useState("");
  const [projectId, setProjectId] = useState("");
  // Default to the whole current month (first to last day).
  const [from, setFrom] = useState(fmtDay(new Date(now.getFullYear(), now.getMonth(), 1)));
  const [to, setTo] = useState(fmtDay(new Date(now.getFullYear(), now.getMonth() + 1, 0)));
  const [text, setText] = useState("");
  const [half, setHalf] = useState(false);

  const [preview, setPreview] = useState(null);
  const [loadingPreview, setLoadingPreview] = useState(false);
  const [filling, setFilling] = useState(false);

  const clientOptions = useMemo(() => clients.map((c) => ({ value: c.id, label: c.name })), [clients]);
  const projectOptions = useMemo(
    () => projects.filter((p) => p.clientId === clientId).map((p) => ({ value: p.id, label: p.name })),
    [projects, clientId]
  );

  // A project belongs to one client, so changing the client drops a project
  // that no longer fits.
  useEffect(() => {
    if (projectId && !projects.some((p) => p.id === projectId && p.clientId === clientId)) {
      setProjectId("");
    }
  }, [clientId, projectId, projects]);

  // Preview as soon as there's a client and a valid range (debounced). The
  // project is optional for the preview - it only adds the "already logged"
  // flagging - so the counts show before a project is picked.
  useEffect(() => {
    if (!show || !currentOrgId || !clientId || !from || !to || from > to) {
      setPreview(null);
      return;
    }
    let cancelled = false;
    setLoadingPreview(true);
    const timer = setTimeout(async () => {
      try {
        const data = await dispatch(workingDaysApi(currentOrgId, clientId, { from, to, projectId }, user.token));
        if (!cancelled) setPreview(data);
      } catch (e) {
        if (!cancelled) setPreview(null);
      } finally {
        if (!cancelled) setLoadingPreview(false);
      }
    }, 300);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [show, currentOrgId, clientId, projectId, from, to, user.token, dispatch]);

  const alreadyLogged = preview?.alreadyLogged?.length || 0;
  const workingTotal = preview?.workingDays?.length || 0;
  const toCreate = projectId ? Math.max(0, workingTotal - alreadyLogged) : workingTotal;
  const canCreate = !!clientId && !!projectId && from <= to && toCreate > 0 && !filling;

  const handleCreate = async () => {
    if (!canCreate) return;
    setFilling(true);
    try {
      const result = await dispatch(
        fillWorkingDaysApi(
          { clientId, projectId, from, to, text: text.trim() || undefined, half },
          currentOrgId,
          user.token
        )
      );
      toast.success(t("timeTracker.fill.created", { count: result?.created?.length ?? 0 }), toastOptions);
      dispatch(getTasksApi(currentOrgId, user.token, 1));
      onClose();
    } catch (e) {
      toast.error(apiErrorMessage(e, locale), toastOptions);
    } finally {
      setFilling(false);
    }
  };

  const footer = (
    <>
      <button type="button" className="cw-button cw-button--secondary" onClick={onClose} disabled={filling}>
        {t("common.cancel")}
      </button>
      <button type="button" className="cw-button" onClick={handleCreate} disabled={!canCreate}>
        {filling ? t("timeTracker.fill.creating") : t("timeTracker.fill.create", { count: toCreate })}
      </button>
    </>
  );

  return (
    <Modal show={show} title={t("timeTracker.fillWorkingDays")} onClose={onClose} footer={footer} wide>
      <div className={styles.grid}>
        <div className={styles.row}>
          <AutocompleteSelect
            className={styles.select}
            label={t("timeTracker.fill.client")}
            placeholder={t("timeTracker.fill.client")}
            options={clientOptions}
            value={clientId}
            onChange={setClientId}
          />
          <AutocompleteSelect
            className={styles.select}
            label={t("timeTracker.fill.project")}
            placeholder={t("timeTracker.fill.project")}
            options={projectOptions}
            value={projectId}
            onChange={setProjectId}
            disabled={!clientId}
          />
        </div>
        <div className={styles.row}>
          <label className={styles.field}>
            <span>{t("timeTracker.fill.from")}</span>
            <input className="cw-input" type="date" value={from} max={to || undefined} onChange={(e) => setFrom(e.target.value)} />
          </label>
          <label className={styles.field}>
            <span>{t("timeTracker.fill.to")}</span>
            <input className="cw-input" type="date" value={to} min={from || undefined} onChange={(e) => setTo(e.target.value)} />
          </label>
        </div>
        <label className={styles.field}>
          <span>{t("timeTracker.fill.text")}</span>
          <input
            className="cw-input"
            type="text"
            value={text}
            placeholder={t("timeTracker.fill.textPlaceholder")}
            onChange={(e) => setText(e.target.value)}
          />
        </label>
        <label className={styles.checkbox}>
          <input type="checkbox" checked={half} onChange={(e) => setHalf(e.target.checked)} />
          <span>{t("timeTracker.fill.half")}</span>
        </label>
      </div>

      <div className={styles.preview}>
        {loadingPreview ? (
          <p className={styles.muted}>{t("common.loading")}</p>
        ) : !clientId ? (
          <p className={styles.muted}>{t("timeTracker.fill.pickClient")}</p>
        ) : preview ? (
          <>
            <p className={styles.summary}>
              {t("timeTracker.fill.willCreate", { count: toCreate })}
              {preview.country ? ` · ${t("timeTracker.fill.country", { country: preview.country })}` : ""}
            </p>
            <ul className={styles.breakdown}>
              <li>{t("timeTracker.fill.weekendsSkipped", { count: preview.weekends?.length || 0 })}</li>
              <li>{t("timeTracker.fill.holidaysSkipped", { count: preview.holidays?.length || 0 })}</li>
              {projectId && <li>{t("timeTracker.fill.alreadyLogged", { count: alreadyLogged })}</li>}
            </ul>
            {preview.holidays?.length > 0 && (
              <p className={styles.holidays}>
                {preview.holidays.map((h) => `${h.day} ${h.name}`).join(" · ")}
              </p>
            )}
            {!projectId && <p className={styles.muted}>{t("timeTracker.fill.pickProject")}</p>}
          </>
        ) : null}
      </div>
    </Modal>
  );
};

export default FillWorkingDaysModal;
