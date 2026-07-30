import React, { useCallback, useEffect, useMemo, useState } from "react";
import { Alert, FlatList, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import Button from "../../components/Button";
import SelectField from "../../components/SelectField";
import FormField from "../../components/FormField";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import projectLabel from "../../common/projectLabel";
import { groupByDay, formatHMS, padTimeString, toDayString, toHMS } from "../../common/format";
import { listProjectsApi } from "../../redux/projects/projects.actions";
import { listClientsApi } from "../../redux/clients/clients.actions";
import {
  listTimeEntriesApi,
  createTimeEntryApi,
  deleteTimeEntryApi,
} from "../../redux/timeEntries/timeEntries.actions";
import { loadTimer, saveTimer, clearTimer } from "../../storage/timer";

// TimeTrackerScreen is the default tab: a running-timer header (start/stop,
// mirrors cwclock-ui's TaskInput.jsx) above a day-grouped list of entries
// (mirrors TasksApp.jsx's dayGroups). All-day records are created from
// AllDayRecordScreen (the "+" button); tapping a row opens EditRecordScreen.
const TimeTrackerScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: projects } = useSelector((state) => state.projects);
  const { items: clients } = useSelector((state) => state.clients);
  const { items: entries, isLoading, isLoadingMore, hasMore, page } = useSelector((state) => state.timeEntries);

  const [projectId, setProjectId] = useState("");
  const [text, setText] = useState("");
  const [timer, setTimer] = useState(null);
  const [elapsed, setElapsed] = useState(0);
  const [starting, setStarting] = useState(false);

  useEffect(() => {
    if (!orgId) return;
    dispatch(listProjectsApi(orgId));
    dispatch(listClientsApi(orgId));
    dispatch(listTimeEntriesApi(orgId, 1));
    loadTimer().then(setTimer);
  }, [orgId, dispatch]);

  useEffect(() => {
    if (!timer) return undefined;
    const tick = () => setElapsed(Math.max(0, Math.floor((Date.now() - new Date(timer.startedAt).getTime()) / 1000)));
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, [timer]);

  const projectItems = useMemo(
    () => projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) })),
    [projects, clients]
  );

  const handleStart = async () => {
    if (!projectId) return;
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;
    const next = {
      startedAt: new Date().toISOString(),
      projectId: project.id,
      clientId: project.clientId,
      text: text.trim() || project.name,
    };
    setStarting(true);
    await saveTimer(next);
    setTimer(next);
    setStarting(false);
  };

  const handleStop = async () => {
    if (!timer) return;
    const started = new Date(timer.startedAt);
    const now = new Date();
    try {
      await dispatch(
        createTimeEntryApi(orgId, {
          text: timer.text,
          day: toDayString(started),
          start: toHMS(started),
          end: toHMS(now),
          allDay: false,
          half: false,
          clientId: timer.clientId,
          projectId: timer.projectId,
        })
      );
      await clearTimer();
      setTimer(null);
      setText("");
      setProjectId("");
    } catch (e) {
      Alert.alert(t("common.retry"), apiErrorMessage(e, locale));
    }
  };

  const handleDelete = (item) => {
    Alert.alert(t("timeTracker.deleteRecordTitle"), t("timeTracker.deleteRecordBody", { text: item.text }), [
      { text: t("common.cancel"), style: "cancel" },
      { text: t("common.delete"), style: "destructive", onPress: () => dispatch(deleteTimeEntryApi(orgId, item.id)) },
    ]);
  };

  const rows = useMemo(() => {
    const out = [];
    for (const group of groupByDay(entries)) {
      out.push({ rowType: "header", key: `h-${group.day}`, day: group.day, totalSecs: group.totalSecs });
      for (const item of group.items) {
        out.push({ rowType: "entry", key: item.id, item });
      }
    }
    return out;
  }, [entries]);

  const timeLabel = useCallback(
    (item) => {
      if (item.allDay) return t("timeTracker.allDay");
      if (item.half) return t("timeTracker.half");
      return `${padTimeString(item.start) || "?"} - ${padTimeString(item.end) || "?"}`;
    },
    [t]
  );

  const renderItem = ({ item: row }) => {
    if (row.rowType === "header") {
      return (
        <View style={styles.dayHeader}>
          <Text style={styles.dayHeaderText}>{row.day}</Text>
          <Text style={styles.dayHeaderText}>{formatHMS(row.totalSecs)}</Text>
        </View>
      );
    }
    const item = row.item;
    const project = projects.find((p) => p.id === item.projectId);
    return (
      <TouchableOpacity style={styles.entryRow} onPress={() => navigation.navigate("EditRecord", { entry: item })}>
        <View style={styles.entryMain}>
          <Text style={styles.entryText} numberOfLines={1}>
            {item.text}
          </Text>
          <Text style={styles.entrySub} numberOfLines={1}>
            {project ? projectLabel(project, clients) : ""} · {timeLabel(item)}
          </Text>
        </View>
        <TouchableOpacity onPress={() => handleDelete(item)} style={styles.deleteBtn}>
          <Text style={styles.deleteBtnText}>{t("common.delete")}</Text>
        </TouchableOpacity>
      </TouchableOpacity>
    );
  };

  return (
    <Screen scroll={false}>
      <View style={styles.header}>
        {timer ? (
          <>
            <Text style={styles.runningText} numberOfLines={1}>
              {timer.text}
            </Text>
            <Text style={styles.elapsed}>{formatHMS(elapsed)}</Text>
            <Button title={t("timeTracker.stop")} variant="danger" onPress={handleStop} />
          </>
        ) : (
          <>
            <SelectField
              label={t("timeTracker.project")}
              value={projectId}
              onValueChange={setProjectId}
              items={projectItems}
              placeholder={t("timeTracker.project")}
            />
            <FormField
              placeholder={t("timeTracker.whatAreYouWorkingOn")}
              value={text}
              onChangeText={setText}
            />
            <Button
              title={t("timeTracker.start")}
              onPress={handleStart}
              disabled={!projectId}
              loading={starting}
            />
          </>
        )}
      </View>

      <FlatList
        data={rows}
        keyExtractor={(row) => row.key}
        renderItem={renderItem}
        refreshing={isLoading}
        onRefresh={() => dispatch(listTimeEntriesApi(orgId, 1))}
        onEndReachedThreshold={0.4}
        onEndReached={() => {
          if (hasMore && !isLoadingMore) dispatch(listTimeEntriesApi(orgId, page + 1));
        }}
        ListEmptyComponent={!isLoading ? <Text style={styles.empty}>{t("timeTracker.noRecords")}</Text> : null}
        contentContainerStyle={styles.listContent}
      />

      <TouchableOpacity style={styles.fab} onPress={() => navigation.navigate("AllDayRecord")}>
        <Text style={styles.fabText}>+</Text>
      </TouchableOpacity>
    </Screen>
  );
};

const styles = StyleSheet.create({
  header: {
    padding: theme.spacing(2),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  runningText: {
    fontSize: 16,
    fontWeight: "600",
    color: theme.color.text,
  },
  elapsed: {
    fontSize: 28,
    fontWeight: "700",
    color: theme.color.primary,
    marginVertical: theme.spacing(1),
  },
  listContent: {
    paddingBottom: theme.spacing(10),
  },
  dayHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    paddingHorizontal: theme.spacing(2),
    paddingVertical: theme.spacing(1),
    backgroundColor: theme.color.backgroundMuted,
  },
  dayHeaderText: {
    fontSize: 13,
    fontWeight: "600",
    color: theme.color.textMuted,
  },
  entryRow: {
    flexDirection: "row",
    alignItems: "center",
    paddingHorizontal: theme.spacing(2),
    paddingVertical: theme.spacing(1.5),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  entryMain: {
    flex: 1,
  },
  entryText: {
    fontSize: 16,
    color: theme.color.text,
  },
  entrySub: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: 2,
  },
  deleteBtn: {
    paddingHorizontal: theme.spacing(1),
  },
  deleteBtnText: {
    color: theme.color.danger,
    fontSize: 13,
  },
  empty: {
    padding: theme.spacing(3),
    textAlign: "center",
    color: theme.color.textMuted,
  },
  fab: {
    position: "absolute",
    right: theme.spacing(2),
    bottom: theme.spacing(2),
    width: 56,
    height: 56,
    borderRadius: 28,
    backgroundColor: theme.color.primary,
    alignItems: "center",
    justifyContent: "center",
    elevation: 4,
  },
  fabText: {
    color: theme.color.white,
    fontSize: 28,
    lineHeight: 30,
  },
});

export default TimeTrackerScreen;
