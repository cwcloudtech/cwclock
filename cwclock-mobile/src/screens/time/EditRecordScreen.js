import React, { useEffect, useMemo, useState } from "react";
import { Alert, StyleSheet } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import DateField from "../../components/DateField";
import SelectField from "../../components/SelectField";
import ToggleRow from "../../components/ToggleRow";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import projectLabel from "../../common/projectLabel";
import { padTimeString, toDayString, toHMS } from "../../common/format";
import { listProjectsApi } from "../../redux/projects/projects.actions";
import { updateTimeEntryApi, deleteTimeEntryApi } from "../../redux/timeEntries/timeEntries.actions";

const timeStringToDate = (hms) => {
  const [h = 9, m = 0, s = 0] = (padTimeString(hms) || "09:00:00").split(":").map(Number);
  const d = new Date();
  d.setHours(h, m, s, 0);
  return d;
};

// EditRecordScreen mirrors cwclock-ui's TaskComponent.jsx/ReportEntryRow.jsx
// edit form: text/day/project plus All day / Half day toggles that disable
// each other (they're mutually exclusive server-side, see ai-instruct-104),
// with start/end time fields only shown when neither is set.
const EditRecordScreen = ({ route, navigation }) => {
  const { entry } = route.params;
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: projects } = useSelector((state) => state.projects);
  const { items: clients } = useSelector((state) => state.clients);

  const [text, setText] = useState(entry.text);
  const [day, setDay] = useState(() => new Date(`${entry.day}T00:00:00`));
  const [projectId, setProjectId] = useState(entry.projectId);
  const [allDay, setAllDay] = useState(entry.allDay);
  const [half, setHalf] = useState(entry.half);
  const [start, setStart] = useState(() => timeStringToDate(entry.start));
  const [end, setEnd] = useState(() => timeStringToDate(entry.end));
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (orgId) dispatch(listProjectsApi(orgId));
  }, [orgId, dispatch]);

  const projectItems = useMemo(
    () => projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) })),
    [projects, clients]
  );

  const handleSave = async () => {
    const project = projects.find((p) => p.id === projectId);
    setError(null);
    setSaving(true);
    try {
      await dispatch(
        updateTimeEntryApi(orgId, {
          ...entry,
          text,
          day: toDayString(day),
          allDay,
          half,
          start: allDay || half ? null : toHMS(start),
          end: allDay || half ? null : toHMS(end),
          projectId,
          clientId: project ? project.clientId : entry.clientId,
        })
      );
      navigation.goBack();
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = () => {
    Alert.alert(t("timeTracker.deleteRecordTitle"), t("timeTracker.deleteRecordBody", { text: entry.text }), [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("common.delete"),
        style: "destructive",
        onPress: async () => {
          await dispatch(deleteTimeEntryApi(orgId, entry.id));
          navigation.goBack();
        },
      },
    ]);
  };

  return (
    <Screen>
      <FormField label={t("timeTracker.taskDescription")} value={text} onChangeText={setText} />
      <SelectField
        label={t("timeTracker.project")}
        value={projectId}
        onValueChange={setProjectId}
        items={projectItems}
      />
      <DateField label={t("timeTracker.day")} value={day} onChange={setDay} mode="date" />
      <ToggleRow
        label={t("timeTracker.allDay")}
        value={allDay}
        disabled={half}
        onValueChange={setAllDay}
      />
      <ToggleRow
        label={t("timeTracker.half")}
        value={half}
        disabled={allDay}
        onValueChange={setHalf}
      />
      {!allDay && !half && (
        <>
          <DateField label={t("timeTracker.startTime")} value={start} onChange={setStart} mode="time" />
          <DateField label={t("timeTracker.endTime")} value={end} onChange={setEnd} mode="time" />
        </>
      )}
      <ErrorBanner message={error} />
      <Button title={t("common.save")} onPress={handleSave} loading={saving} style={styles.saveButton} />
      <Button title={t("common.delete")} variant="danger" onPress={handleDelete} />
    </Screen>
  );
};

const styles = StyleSheet.create({
  saveButton: {
    marginBottom: theme.spacing(1.5),
  },
});

export default EditRecordScreen;
