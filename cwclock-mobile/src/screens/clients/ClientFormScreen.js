import React, { useEffect, useMemo, useState } from "react";
import { Alert, StyleSheet } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import SelectField from "../../components/SelectField";
import ToggleRow from "../../components/ToggleRow";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import { listCountriesApi } from "../../redux/countries/countries.actions";
import { createClientApi, updateClientApi, deleteClientApi } from "../../redux/clients/clients.actions";

const emptyFields = {
  name: "",
  email: "",
  contactName: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  vatRate: "",
  identificationNumber: "",
  purchaseOrder: "",
  hoursPerDay: "",
  dailyRate: "",
  sendReportsWithInvoice: false,
};

// ClientFormScreen: create (route.params.client absent) or edit (present).
// Ported from cwclock-ui's EditClientModal.jsx, trimmed to the core fields -
// no country-specific identification fields (SIREN/SIRET/NAF/MF, see
// useCountryFields.js - a generic "Identification number" field covers that
// need here), invoiceEmails/vatDischargeMotive, or cross-organization
// transfer (ai-instruct-106 scope).
const ClientFormScreen = ({ route, navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: countries } = useSelector((state) => state.countries);
  const editingClient = route.params?.client;

  const [fields, setFields] = useState(() =>
    editingClient
      ? {
          name: editingClient.name || "",
          email: editingClient.email || "",
          contactName: editingClient.contactName || "",
          address: editingClient.address || "",
          postalCode: editingClient.postalCode || "",
          city: editingClient.city || "",
          country: editingClient.country || "",
          vatNumber: editingClient.vatNumber || "",
          vatRate: editingClient.vatRate ? String(editingClient.vatRate) : "",
          identificationNumber: editingClient.identificationNumber || "",
          purchaseOrder: editingClient.purchaseOrder || "",
          hoursPerDay: editingClient.hoursPerDay ? String(editingClient.hoursPerDay) : "",
          dailyRate: editingClient.dailyRate != null ? String(editingClient.dailyRate) : "",
          sendReportsWithInvoice: editingClient.sendReportsWithInvoice || false,
        }
      : emptyFields
  );
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    dispatch(listCountriesApi());
  }, [dispatch]);

  const countryItems = useMemo(() => countries.map((c) => ({ value: c.iso, label: c.name })), [countries]);

  const setField = (key, value) => setFields((f) => ({ ...f, [key]: value }));

  const toNumber = (v) => (v === "" ? undefined : Number(v));

  const handleSave = async () => {
    setError(null);
    setSaving(true);
    const payload = {
      ...fields,
      vatRate: toNumber(fields.vatRate),
      hoursPerDay: toNumber(fields.hoursPerDay),
      dailyRate: toNumber(fields.dailyRate),
    };
    try {
      if (editingClient) {
        await dispatch(updateClientApi(orgId, editingClient.id, payload));
      } else {
        await dispatch(createClientApi(orgId, payload));
      }
      navigation.goBack();
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = () => {
    Alert.alert(t("clients.deleteClientTitle"), t("clients.deleteClientBody", { name: editingClient.name }), [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("common.delete"),
        style: "destructive",
        onPress: async () => {
          await dispatch(deleteClientApi(orgId, editingClient.id));
          navigation.goBack();
        },
      },
    ]);
  };

  return (
    <Screen>
      <FormField label={t("clients.name")} value={fields.name} onChangeText={(v) => setField("name", v)} />
      <FormField
        label={t("clients.email")}
        value={fields.email}
        onChangeText={(v) => setField("email", v)}
        autoCapitalize="none"
        keyboardType="email-address"
      />
      <FormField
        label={t("clients.contactName")}
        value={fields.contactName}
        onChangeText={(v) => setField("contactName", v)}
      />
      <FormField label={t("clients.address")} value={fields.address} onChangeText={(v) => setField("address", v)} />
      <FormField
        label={t("clients.postalCode")}
        value={fields.postalCode}
        onChangeText={(v) => setField("postalCode", v)}
      />
      <FormField label={t("clients.city")} value={fields.city} onChangeText={(v) => setField("city", v)} />
      <SelectField
        label={t("clients.country")}
        value={fields.country}
        onValueChange={(v) => setField("country", v)}
        items={countryItems}
        placeholder={t("clients.country")}
      />
      <FormField
        label={t("clients.vatNumber")}
        value={fields.vatNumber}
        onChangeText={(v) => setField("vatNumber", v)}
      />
      <FormField
        label={t("clients.vatRate")}
        value={fields.vatRate}
        onChangeText={(v) => setField("vatRate", v)}
        keyboardType="decimal-pad"
      />
      <FormField
        label={t("clients.identificationNumber")}
        value={fields.identificationNumber}
        onChangeText={(v) => setField("identificationNumber", v)}
      />
      <FormField
        label={t("clients.purchaseOrder")}
        value={fields.purchaseOrder}
        onChangeText={(v) => setField("purchaseOrder", v)}
      />
      <FormField
        label={t("clients.hoursPerDay")}
        value={fields.hoursPerDay}
        onChangeText={(v) => setField("hoursPerDay", v)}
        keyboardType="decimal-pad"
      />
      <FormField
        label={t("clients.dailyRate")}
        value={fields.dailyRate}
        onChangeText={(v) => setField("dailyRate", v)}
        keyboardType="decimal-pad"
      />
      <ToggleRow
        label={t("clients.sendReportsWithInvoice")}
        value={fields.sendReportsWithInvoice}
        onValueChange={(v) => setField("sendReportsWithInvoice", v)}
      />
      <ErrorBanner message={error} />
      <Button
        title={t("common.save")}
        onPress={handleSave}
        loading={saving}
        disabled={!fields.name.trim() || !fields.country}
        style={editingClient ? styles.saveButton : undefined}
      />
      {editingClient && <Button title={t("common.delete")} variant="danger" onPress={handleDelete} />}
    </Screen>
  );
};

const styles = StyleSheet.create({
  saveButton: {
    marginBottom: theme.spacing(1.5),
  },
});

export default ClientFormScreen;
