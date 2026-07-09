import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import { FaChartBar } from "react-icons/fa";
import { useI18n } from "../../i18n/I18nContext";
import DateRangePicker from "../common/DateRangePicker";
import MultiSelect from "../common/MultiSelect";
import Dropdown, { DropdownItem } from "../common/Dropdown";
import Spinner from "../spinner/Spinner";
import memberLabel from "../common/memberLabel";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { listMembersApi } from "../../Redux/Organizations/Org.actions";
import { fetchReportApi, exportReportApi } from "../../Redux/Reports/Report.actions";
import { dateRangeShortcuts, toISODate } from "../common/dateRangeShortcuts";
import { isAdminOrOwner as computeIsAdminOrOwner, isSuperadmin, memberRole } from "../common/permissions";
import SummaryReportView from "./SummaryReportView";
import DetailedReportView from "./DetailedReportView";
import styles from "./Styles/Reports.module.css";

const defaultRange = (t) => {
  const thisWeek = dateRangeShortcuts(t).find((s) => s.key === "thisWeek");
  const [s, e] = thisWeek.range();
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

  const setRange = (start, end) => setRangeState({ start, end });

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

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

  const refresh = () => {
    if (currentOrgId && canAccess) {
      dispatch(fetchReportApi(currentOrgId, filters, user.token));
    }
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
