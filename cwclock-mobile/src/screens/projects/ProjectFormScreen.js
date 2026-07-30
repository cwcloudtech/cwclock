import React, { useEffect, useMemo, useState } from "react";
import { Alert, StyleSheet } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import SelectField from "../../components/SelectField";
import ColorSwatchPicker from "../../components/ColorSwatchPicker";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import { listClientsApi } from "../../redux/clients/clients.actions";
import { createProjectApi, updateProjectApi, deleteProjectApi } from "../../redux/projects/projects.actions";

const DEFAULT_COLOR = "#1cb9f7";

const subdivisionsToText = (subdivisions) => (Array.isArray(subdivisions) ? subdivisions.join(", ") : "");
const textToSubdivisions = (text) =>
  text
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

// ProjectFormScreen: create (route.params.project absent) or edit (present).
// Ported from cwclock-ui's EditProjectModal.jsx - color uses a fixed swatch
// row instead of a native color picker (see ColorSwatchPicker.js), and
// cross-client transfer is left to the web app, matching ClientFormScreen's
// equivalent trims.
const ProjectFormScreen = ({ route, navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: clients } = useSelector((state) => state.clients);
  const editingProject = route.params?.project;

  const [name, setName] = useState(editingProject?.name || "");
  const [clientId, setClientId] = useState(editingProject?.clientId || "");
  const [color, setColor] = useState(editingProject?.color || DEFAULT_COLOR);
  const [dailyRate, setDailyRate] = useState(editingProject?.dailyRate != null ? String(editingProject.dailyRate) : "");
  const [subdivisionsText, setSubdivisionsText] = useState(subdivisionsToText(editingProject?.subdivisions));
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (orgId) dispatch(listClientsApi(orgId));
  }, [orgId, dispatch]);

  const clientItems = useMemo(() => clients.map((c) => ({ value: c.id, label: c.name })), [clients]);

  const handleSave = async () => {
    if (!name.trim() || !clientId) return;
    setError(null);
    setSaving(true);
    const fields = {
      name: name.trim(),
      color,
      dailyRate: dailyRate === "" ? undefined : Number(dailyRate),
      subdivisions: textToSubdivisions(subdivisionsText),
    };
    try {
      if (editingProject) {
        await dispatch(updateProjectApi(orgId, editingProject.id, { ...fields, clientId }));
      } else {
        await dispatch(createProjectApi(orgId, clientId, fields));
      }
      navigation.goBack();
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = () => {
    Alert.alert(t("projects.deleteProjectTitle"), t("projects.deleteProjectBody", { name: editingProject.name }), [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("common.delete"),
        style: "destructive",
        onPress: async () => {
          await dispatch(deleteProjectApi(orgId, editingProject.id));
          navigation.goBack();
        },
      },
    ]);
  };

  return (
    <Screen>
      <FormField label={t("projects.name")} value={name} onChangeText={setName} />
      <SelectField
        label={t("projects.client")}
        value={clientId}
        onValueChange={setClientId}
        items={clientItems}
        placeholder={t("projects.client")}
      />
      <ColorSwatchPicker label={t("projects.color")} value={color} onChange={setColor} />
      <FormField
        label={t("projects.dailyRate")}
        value={dailyRate}
        onChangeText={setDailyRate}
        keyboardType="decimal-pad"
      />
      <FormField
        label={t("projects.subdivisions")}
        value={subdivisionsText}
        onChangeText={setSubdivisionsText}
        placeholder={t("projects.subdivisionsHint")}
      />
      <ErrorBanner message={error} />
      <Button
        title={t("common.save")}
        onPress={handleSave}
        loading={saving}
        disabled={!name.trim() || !clientId}
        style={editingProject ? styles.saveButton : undefined}
      />
      {editingProject && <Button title={t("common.delete")} variant="danger" onPress={handleDelete} />}
    </Screen>
  );
};

const styles = StyleSheet.create({
  saveButton: {
    marginBottom: theme.spacing(1.5),
  },
});

export default ProjectFormScreen;
