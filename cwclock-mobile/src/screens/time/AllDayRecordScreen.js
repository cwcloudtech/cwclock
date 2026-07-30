import React, { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import DateField from "../../components/DateField";
import SelectField from "../../components/SelectField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import projectLabel from "../../common/projectLabel";
import { toDayString } from "../../common/format";
import { listProjectsApi } from "../../redux/projects/projects.actions";
import { createTimeEntryApi } from "../../redux/timeEntries/timeEntries.actions";

// AllDayRecordScreen is the explicit "record all day with a date picker"
// feature from the ai-instruct-105 plan - picking a project and a date
// creates the entry immediately (mirrors cwclock-ui's TaskInput.jsx
// handleAllDayDate: no separate "create" step once the date is picked
// there; here Save is explicit since there's no equivalent icon-button
// shortcut on mobile).
const AllDayRecordScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: projects } = useSelector((state) => state.projects);
  const { items: clients } = useSelector((state) => state.clients);

  const [projectId, setProjectId] = useState("");
  const [text, setText] = useState("");
  const [day, setDay] = useState(() => new Date());
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (orgId) dispatch(listProjectsApi(orgId));
  }, [orgId, dispatch]);

  const projectItems = useMemo(
    () => projects.map((p) => ({ value: p.id, label: projectLabel(p, clients) })),
    [projects, clients]
  );

  const handleCreate = async () => {
    const project = projects.find((p) => p.id === projectId);
    if (!project) return;
    setError(null);
    setSaving(true);
    try {
      await dispatch(
        createTimeEntryApi(orgId, {
          text: text.trim() || project.name,
          day: toDayString(day),
          allDay: true,
          half: false,
          clientId: project.clientId,
          projectId: project.id,
        })
      );
      navigation.goBack();
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Screen>
      <SelectField
        label={t("timeTracker.project")}
        value={projectId}
        onValueChange={setProjectId}
        items={projectItems}
        placeholder={t("timeTracker.project")}
      />
      <FormField
        label={t("timeTracker.taskDescription")}
        value={text}
        onChangeText={setText}
        placeholder={t("timeTracker.whatAreYouWorkingOn")}
      />
      <DateField label={t("timeTracker.day")} value={day} onChange={setDay} mode="date" />
      <ErrorBanner message={error} />
      <Button title={t("timeTracker.addAllDayRecord")} onPress={handleCreate} disabled={!projectId} loading={saving} />
    </Screen>
  );
};

export default AllDayRecordScreen;
