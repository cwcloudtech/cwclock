import React, { useState, useMemo, useRef } from "react";
import TaskInput from "./TaskInput";
import { useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import TaskComponent from "./TaskComponent";
import styles from "./Styles/TaskApp.module.css";
import EmptyTask from "./EmptyTask";
import ImportCSVModal from "./ImportCSVModal";
import { useNavigate, Link } from "react-router-dom";
import Spinner from "../spinner/Spinner";
import { getTasksApi } from "../../Redux/Tasks/Task.actions";
import { listOrgsApi, listMembersApi } from "../../Redux/Organizations/Org.actions";
import { listClientsApi } from "../../Redux/Clients/Client.actions";
import { listProjectsApi } from "../../Redux/Projects/Project.actions";
import { formatHMS } from "../Reports/reportFormat";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../common/permissions";
import { useI18n } from "../../i18n/I18nContext";

// Duration math lives here rather than reusing the reports package's
// enrichment logic: that logic pulls in per-client hours-per-day / all-day
// window billing semantics this personal screen doesn't need - an all-day
// entry simply contributes nothing to its day's total here.
const parseSecondsOfDay = (hms) => {
  if (!hms) return 0;
  const [h, m, s] = hms.split(":").map(Number);
  return (h || 0) * 3600 + (m || 0) * 60 + (s || 0);
};

const entryDurationSecs = (item) => {
  if (item.allDay || !item.start || !item.end) return 0;
  const secs = parseSecondsOfDay(item.end) - parseSecondsOfDay(item.start);
  return secs > 0 ? secs : 0;
};

const TasksApp = () => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [showImport, setShowImport] = useState(false);
  const { tasks, isLoading, isLoadingMore, hasMore, page } = useSelector((state) => state.tasks);
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const isAdminOrOwner = computeIsAdminOrOwner(user, members);
  const sentinelRef = useRef(null);

  useEffect(() => {
    if (user.token) {
      dispatch(listOrgsApi(user.token));
    } else {
      navigate("/login");
    }
  }, [user.token, dispatch, navigate]);

  useEffect(() => {
    if (user.token && currentOrgId) {
      dispatch(getTasksApi(currentOrgId, user.token));
      dispatch(listClientsApi(currentOrgId, user.token));
      dispatch(listProjectsApi(currentOrgId, user.token));
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [user.token, currentOrgId, dispatch]);

  // Loads the next page once the sentinel below the list scrolls into view.
  useEffect(() => {
    const el = sentinelRef.current;
    if (!el || !hasMore) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting && !isLoadingMore) {
          dispatch(getTasksApi(currentOrgId, user.token, page + 1));
        }
      },
      { rootMargin: "200px" }
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [hasMore, isLoadingMore, page, currentOrgId, user.token, dispatch]);

  // Entries already arrive sorted day DESC from the API, and offset
  // pagination preserves that order across page boundaries, so a single
  // pass grouping consecutive same-day items is enough - no re-sort needed.
  const dayGroups = useMemo(() => {
    const groups = [];
    for (const item of tasks) {
      const last = groups[groups.length - 1];
      if (last && last.day === item.day) {
        last.items.push(item);
        last.totalSecs += entryDurationSecs(item);
      } else {
        groups.push({ day: item.day, items: [item], totalSecs: entryDurationSecs(item) });
      }
    }
    return groups;
  }, [tasks]);

  if (isLoading) {
    return (
      <div className={styles.Body1}>
        <Spinner />
      </div>
    );
  }

  if (!currentOrgId) {
    return (
      <div className={styles.Body1}>
        <div className={styles.Empty}>
          <p>
            {t("timeTracker.needOrganization")}{" "}
            <Link to="/dashboard/organizations">{t("common.createOne")}</Link>.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={styles.Body1}>
      <TaskInput
        isAdminOrOwner={isAdminOrOwner}
        onImportClick={() => setShowImport(true)}
      />
      {tasks.length <= 0 ? (
        <div>
          <EmptyTask />
        </div>
      ) : (
        <div>
          {dayGroups.map((group) => (
            <React.Fragment key={group.day}>
              <div className={styles.DaySeparator}>
                <span>{group.day}</span>
                <span>{formatHMS(group.totalSecs)}</span>
              </div>
              {group.items.map((item) => (
                <TaskComponent key={item.id} item={item} />
              ))}
            </React.Fragment>
          ))}
          {hasMore && <div ref={sentinelRef} className={styles.Sentinel} />}
          {isLoadingMore && (
            <div className={styles.LoadMoreSpinner}>
              {/* Spinner itself is a fixed full-page overlay (App.css
                  .loadingSpinnerContainer), wrong for an inline bottom-of-
                  list indicator, so only its rotating-circle class is reused
                  here. */}
              <div className="loadingSpinner" />
            </div>
          )}
        </div>
      )}
      <ImportCSVModal show={showImport} onClose={() => setShowImport(false)} />
    </div>
  );
};

export default TasksApp;
