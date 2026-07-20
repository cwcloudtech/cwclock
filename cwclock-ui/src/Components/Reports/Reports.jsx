import React, { useEffect, useRef, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { toast } from "react-toastify";
import { FaChartBar } from "react-icons/fa";
import { useI18n } from "../../i18n/I18nContext";
import DateRangePicker from "../common/DateRangePicker";
import MultiSelect from "../common/MultiSelect";
import Dropdown, { DropdownItem } from "../common/Dropdown";
import Spinner from "../spinner/Spinner";
import memberLabel from "../common/memberLabel";
import toastOptions from "../../Redux/toastOptions";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import { fetchReportApi, exportReportApi, clearReport } from "../../Redux/Reports/Report.actions";
import { fetchLatestEntryDayApi } from "../../Redux/Tasks/Task.actions";
import { dateRangeShortcuts, toISODate, fromISODate } from "../common/dateRangeShortcuts";
import { isAdminOrOwner as computeIsAdminOrOwner, isSuperadmin, memberRole } from "../common/permissions";
import SummaryReportView from "./SummaryReportView";
import DetailedReportView from "./DetailedReportView";
import styles from "./Styles/Reports.module.css";

// referenceDate defaults to today - pass the user's most recently logged
// time entry's date instead once it's known (see the effect below), so the
// initial range reflects their last activity rather than an empty range
// when they haven't logged anything recently (ai-instruct-63). Start is
// always the first day of that reference date's month.
const defaultRange = (t, referenceDate) => {
  const thisMonth = dateRangeShortcuts(t, referenceDate).find((s) => s.key === "thisMonth");
  const [s, e] = thisMonth.range();
  return { start: toISODate(s), end: toISODate(e) };
};

const Reports = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const { projects } = useSelector((state) => state.projects);
  const { data: report, isLoading } = useSelector((state) => state.reports);

  const [tab, setTab] = useState("summary");
  const [range, setRangeState] = useState(() => defaultRange(t));
  const [clientIds, setClientIds] = useState([]);
  const [projectIds, setProjectIds] = useState([]);
  const [userIds, setUserIds] = useState([]);

  const myRole = memberRole(user, members);
  const canAccess = isSuperadmin(user) || (myRole && myRole !== "reader");
  const isAdminOrOwner = computeIsAdminOrOwner(user, members);

  // Tracks whether the user has touched the date range themselves, so the
  // most-recent-entry lookup below never clobbers a choice they already
  // made - a ref rather than state since it's read from inside an async
  // callback (see the effect below), where a state closure would go stale.
  const rangeTouchedRef = useRef(false);
  const setRange = (start, end) => {
    rangeTouchedRef.current = true;
    setRangeState({ start, end });
  };

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  // Defaults the date range to the week of the user's most recently logged
  // time entry instead of the real current date, so it isn't just an empty
  // "this week" for someone who hasn't logged anything in a while
  // (ai-instruct-63). Only applies once, and only if the range is still
  // untouched by the time it resolves.
  useEffect(() => {
    if (!currentOrgId) return;
    let cancelled = false;
    dispatch(fetchLatestEntryDayApi(currentOrgId, user.token)).then((day) => {
      if (cancelled || !day || rangeTouchedRef.current) return;
      setRangeState(defaultRange(t, fromISODate(day)));
    });
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentOrgId]);

  // Narrowing the client filter narrows which projects make sense too, so
  // drop any already-selected project that no longer belongs to one of the
  // selected clients.
  useEffect(() => {
    if (clientIds.length === 0) return;
    setProjectIds((prev) =>
      prev.filter((id) => clientIds.includes(projects.find((p) => p.id === id)?.clientId))
    );
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [clientIds, projects]);

  const filters = { type: tab, start: range.start, end: range.end, clientIds, projectIds, userIds };
  const isSummaryReport = report && Array.isArray(report.rows);
  const isDetailedReport = report && Array.isArray(report.entries);
  const isValidRange = range.start <= range.end;

  const refresh = () => {
    if (!currentOrgId || !canAccess) return;
    if (!isValidRange) {
      toast.error(t("errors.invalidDateRange"), toastOptions);
      dispatch(clearReport());
      return;
    }
    dispatch(fetchReportApi(currentOrgId, filters, user.token));
  };

  useEffect(() => {
    refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentOrgId, canAccess, tab, range.start, range.end, clientIds, projectIds, userIds]);

  if (!currentOrgId) {
    return <h1 className="cw-title">{t("organizations.selectOrCreateFirst")}</h1>;
  }

  if (!canAccess) {
    return <p className="cw-error">{t("reports.noAccess")}</p>;
  }

  const clientOptions = clients.map((c) => ({ value: c.id, label: c.name }));
  const projectOptions = projects
    .filter((p) => clientIds.length === 0 || clientIds.includes(p.clientId))
    .map((p) => ({ value: p.id, label: p.name }));
  const memberOptions = members.map((m) => ({ value: m.userId, label: memberLabel(m) }));

  const handleExport = (format) => {
    if (!isValidRange) {
      toast.error(t("errors.invalidDateRange"), toastOptions);
      return;
    }
    dispatch(exportReportApi(currentOrgId, filters, format, user.token));
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">
        <FaChartBar className={styles.titleIcon} /> {t("nav.reports")}
      </h1>

      <div className={styles.toolbar}>
        <div className={styles.tabs}>
          <button
            type="button"
            className={`${styles.tab} ${tab === "summary" ? styles.tabActive : ""}`}
            onClick={() => setTab("summary")}
          >
            {t("reports.summary")}
          </button>
          <button
            type="button"
            className={`${styles.tab} ${tab === "detailed" ? styles.tabActive : ""}`}
            onClick={() => setTab("detailed")}
          >
            {t("reports.detailed")}
          </button>
        </div>

        <DateRangePicker start={range.start} end={range.end} onChange={setRange} />

        <Dropdown
          title={t("reports.export")}
          align="end"
          triggerClassName={styles.exportTrigger}
          trigger={<>{t("reports.export")}</>}
        >
          {(close) => (
            <>
              <DropdownItem
                onClick={() => {
                  handleExport("csv");
                  close();
                }}
              >
                {t("reports.exportCsv")}
              </DropdownItem>
              <DropdownItem
                onClick={() => {
                  handleExport("pdf");
                  close();
                }}
              >
                {t("reports.exportPdf")}
              </DropdownItem>
            </>
          )}
        </Dropdown>
      </div>

      <div className={styles.filters}>
        <MultiSelect label={t("common.client")} options={clientOptions} selected={clientIds} onChange={setClientIds} />
        <MultiSelect label={t("projects.title")} options={projectOptions} selected={projectIds} onChange={setProjectIds} />
        <MultiSelect label={t("nav.users")} options={memberOptions} selected={userIds} onChange={setUserIds} />
      </div>

      {isLoading && <Spinner />}
      {/* report can still hold the other tab's shape for one render after
          switching tabs, until the new fetch resolves — check the actual
          shape, not just tab === report's presence, or the wrong view can
          be handed the wrong-shaped data and crash. */}
      {!isLoading && isSummaryReport && tab === "summary" && <SummaryReportView report={report} />}
      {!isLoading && isDetailedReport && tab === "detailed" && (
        <DetailedReportView report={report} orgId={currentOrgId} isAdminOrOwner={isAdminOrOwner} onChanged={refresh} />
      )}
    </div>
  );
};

export default Reports;
