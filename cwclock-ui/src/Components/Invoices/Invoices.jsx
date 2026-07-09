import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaFileInvoiceDollar, FaRegEdit } from "react-icons/fa";
import { FiDownload } from "react-icons/fi";
import { useI18n } from "../../i18n/I18nContext";
import DateRangePicker from "../common/DateRangePicker";
import Spinner from "../spinner/Spinner";
import EmptyState from "../common/EmptyState";
import Button from "../common/Button";
import Tooltip from "../common/Tooltip";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import {
  listInvoicesApi,
  previewInvoiceApi,
  generateInvoiceApi,
  downloadInvoicePdfApi,
  updateInvoiceStatusApi,
} from "../../Redux/Invoices/Invoice.actions";
import { dateRangeShortcuts, toISODate } from "../common/dateRangeShortcuts";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../common/permissions";
import { apiErrorMessage } from "../../i18n/translate";
import styles from "./Styles/Invoices.module.css";

// Invoices are generated a month at a time in practice, so the picker only
// offers those two shortcuts rather than the full report shortcut list.
const INVOICE_SHORTCUT_KEYS = ["thisMonth", "lastMonth"];

const STATUSES = ["unpaid", "paid", "canceled", "refunded"];
const STATUS_LABEL_KEY = {
  unpaid: "invoices.statusUnpaid",
  paid: "invoices.statusPaid",
  canceled: "invoices.statusCanceled",
  refunded: "invoices.statusRefunded",
};

const defaultRange = (t) => {
  const thisMonth = dateRangeShortcuts(t).find((s) => s.key === "thisMonth");
  const [s, e] = thisMonth.range();
  return { start: toISODate(s), end: toISODate(e) };
};

const InvoiceRow = ({ invoice, orgId, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [editing, setEditing] = useState(false);
  const [status, setStatus] = useState(invoice.status);

  const handleSave = async () => {
    await dispatch(updateInvoiceStatusApi(orgId, invoice.id, status, token));
    setEditing(false);
  };

  return (
    <li className={`cw-list-item ${styles.invoiceRow}`}>
      <span className={styles.number}>{invoice.number}</span>
      <span className={styles.period}>
        {invoice.selectedBeginDate} - {invoice.selectedEndDate}
      </span>
      {editing ? (
        <select className="cw-select" value={status} onChange={(e) => setStatus(e.target.value)} title={t("invoices.status")}>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {t(STATUS_LABEL_KEY[s])}
            </option>
          ))}
        </select>
      ) : (
        <span className={styles.status}>{t(STATUS_LABEL_KEY[invoice.status])}</span>
      )}
      <span className={styles.total}>{invoice.totalTTC.toFixed(2)}</span>
      <div className={styles.rowActions}>
        <Tooltip label={t("invoices.download")}>
          <button
            type="button"
            className={styles.iconBtn}
            onClick={() => dispatch(downloadInvoicePdfApi(orgId, invoice.id, token))}
          >
            <FiDownload style={{ fontSize: "16px" }} />
          </button>
        </Tooltip>
        {editing ? (
          <Button size="sm" onClick={handleSave} title={t("invoices.saveStatus")}>
            {t("common.save")}
          </Button>
        ) : (
          <Tooltip label={t("invoices.editStatus")}>
            <button type="button" className={styles.iconBtn} onClick={() => setEditing(true)}>
              <FaRegEdit style={{ fontSize: "16px" }} />
            </button>
          </Tooltip>
        )}
      </div>
    </li>
  );
};

const Invoices = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { invoices, isLoading } = useSelector((state) => state.invoices);

  const [range, setRangeState] = useState(() => defaultRange(t));
  const [clientId, setClientId] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const isAdminOrOwner = computeIsAdminOrOwner(user, members);
  const setRange = (start, end) => setRangeState({ start, end });

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const refreshInvoices = () => {
    if (currentOrgId && clientId && isAdminOrOwner) {
      dispatch(listInvoicesApi(currentOrgId, clientId, range.start, range.end, user.token));
    }
  };

  useEffect(() => {
    refreshInvoices();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentOrgId, clientId, range.start, range.end, isAdminOrOwner]);

  if (!currentOrgId) {
    return <h1 className="cw-title">{t("organizations.selectOrCreateFirst")}</h1>;
  }

  if (!isAdminOrOwner) {
    return <p className="cw-error">{t("invoices.noAccess")}</p>;
  }

  const runAction = async (apiThunk) => {
    setError("");
    if (!clientId) {
      setError(t("invoices.clientRequired"));
      return;
    }
    setBusy(true);
    try {
      await dispatch(apiThunk(currentOrgId, clientId, range.start, range.end, user.token));
      refreshInvoices();
    } catch (err) {
      setError(apiErrorMessage(err, locale));
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">
        <FaFileInvoiceDollar className={styles.titleIcon} /> {t("nav.invoices")}
      </h1>

      <div className={styles.toolbar}>
        <select
          className="cw-select"
          value={clientId}
          onChange={(e) => setClientId(e.target.value)}
          title={t("invoices.selectAClient")}
        >
          <option value="">{t("invoices.selectAClient")}</option>
          {clients.map((c) => (
            <option key={c.id} value={c.id}>
              {c.name}
            </option>
          ))}
        </select>

        <DateRangePicker start={range.start} end={range.end} onChange={setRange} shortcutKeys={INVOICE_SHORTCUT_KEYS} />
      </div>

      <div className={styles.actions}>
        <Button variant="secondary" disabled={busy} onClick={() => runAction(previewInvoiceApi)} title={t("invoices.previewHint")}>
          {t("invoices.preview")}
        </Button>
        <Button disabled={busy} onClick={() => runAction(generateInvoiceApi)} title={t("invoices.generateHint")}>
          {t("invoices.generate")}
        </Button>
      </div>
      {error && <p className="cw-error">{error}</p>}

      <h2 className="cw-subtitle">{t("invoices.listTitle")}</h2>
      {isLoading && <Spinner />}
      {!isLoading && clientId && invoices.length === 0 && (
        <EmptyState title={t("invoices.emptyTitle")} body={t("invoices.emptyBody")} />
      )}
      {!isLoading && invoices.length > 0 && (
        <ul className="cw-list">
          {invoices.map((invoice) => (
            <InvoiceRow key={invoice.id} invoice={invoice} orgId={currentOrgId} token={user.token} />
          ))}
        </ul>
      )}
    </div>
  );
};

export default Invoices;
