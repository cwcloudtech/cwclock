import React, { useEffect, useMemo, useState } from "react";
import { Alert, FlatList, StyleSheet, Text, TouchableOpacity, View } from "react-native";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import DateField from "../../components/DateField";
import SelectField from "../../components/SelectField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import theme from "../../theme";
import { toDayString } from "../../common/format";
import { listClientsApi } from "../../redux/clients/clients.actions";
import {
  listInvoicesApi,
  previewInvoicePdfApi,
  generateInvoicePdfApi,
  downloadInvoicePdfApi,
} from "../../redux/invoices/invoices.actions";

const firstOfMonth = () => {
  const d = new Date();
  d.setDate(1);
  return d;
};

// InvoicesScreen: this tab is only registered when the connected user is
// admin/owner of the current org (see App.js, mirroring cwclock-ui's
// Slidebar.jsx showInvoices gate). Preview never saves anything; Generate
// allocates a real invoice number, so it's confirmed first.
const InvoicesScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);
  const { items: clients } = useSelector((state) => state.clients);
  const { items: invoices } = useSelector((state) => state.invoices);

  const [clientId, setClientId] = useState("");
  const [start, setStart] = useState(firstOfMonth);
  const [end, setEnd] = useState(() => new Date());
  const [error, setError] = useState(null);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (orgId) dispatch(listClientsApi(orgId));
  }, [orgId, dispatch]);

  const range = useMemo(() => ({ start: toDayString(start), end: toDayString(end) }), [start, end]);

  useEffect(() => {
    if (orgId && clientId) dispatch(listInvoicesApi(orgId, clientId, range));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [orgId, clientId, range.start, range.end, dispatch]);

  const clientItems = useMemo(() => clients.map((c) => ({ value: c.id, label: c.name })), [clients]);

  const openPdf = (path, title) => navigation.navigate("PdfViewer", { path, title });

  const handlePreview = async () => {
    if (!clientId) return;
    setError(null);
    setBusy(true);
    try {
      const path = await dispatch(previewInvoicePdfApi(orgId, { clientId, ...range }));
      openPdf(path, t("invoices.preview"));
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setBusy(false);
    }
  };

  const handleGenerate = () => {
    if (!clientId) return;
    Alert.alert(t("invoices.generateConfirmTitle"), t("invoices.generateConfirmBody"), [
      { text: t("common.cancel"), style: "cancel" },
      {
        text: t("invoices.generate"),
        onPress: async () => {
          setError(null);
          setBusy(true);
          try {
            const path = await dispatch(generateInvoicePdfApi(orgId, { clientId, ...range }));
            dispatch(listInvoicesApi(orgId, clientId, range));
            openPdf(path, t("invoices.title"));
          } catch (e) {
            setError(apiErrorMessage(e, locale));
          } finally {
            setBusy(false);
          }
        },
      },
    ]);
  };

  const handleOpenInvoice = async (invoice) => {
    try {
      const path = await dispatch(downloadInvoicePdfApi(orgId, invoice.id));
      openPdf(path, invoice.number);
    } catch (e) {
      Alert.alert(t("common.retry"), apiErrorMessage(e, locale));
    }
  };

  return (
    <Screen scroll={false}>
      <View style={styles.form}>
        <SelectField
          label={t("invoices.client")}
          value={clientId}
          onValueChange={setClientId}
          items={clientItems}
          placeholder={t("invoices.client")}
        />
        <DateField label={t("invoices.startDate")} value={start} onChange={setStart} mode="date" />
        <DateField label={t("invoices.endDate")} value={end} onChange={setEnd} mode="date" />
        <ErrorBanner message={error} />
        <View style={styles.buttonRow}>
          <Button
            title={t("invoices.preview")}
            variant="secondary"
            onPress={handlePreview}
            disabled={!clientId}
            loading={busy}
            style={styles.buttonHalf}
          />
          <Button
            title={t("invoices.generate")}
            onPress={handleGenerate}
            disabled={!clientId}
            loading={busy}
            style={styles.buttonHalf}
          />
        </View>
      </View>

      <Text style={styles.sectionTitle}>{t("invoices.previousInvoices")}</Text>
      <FlatList
        data={invoices}
        keyExtractor={(item) => item.id}
        contentContainerStyle={styles.list}
        ListEmptyComponent={<Text style={styles.empty}>{t("invoices.noInvoices")}</Text>}
        renderItem={({ item }) => (
          <TouchableOpacity style={styles.row} onPress={() => handleOpenInvoice(item)}>
            <Text style={styles.rowText}>{item.number}</Text>
            <Text style={styles.rowSub}>
              {item.status} · {Number(item.totalTTC).toFixed(2)}
            </Text>
          </TouchableOpacity>
        )}
      />
    </Screen>
  );
};

const styles = StyleSheet.create({
  form: {
    padding: theme.spacing(2),
  },
  buttonRow: {
    flexDirection: "row",
    gap: theme.spacing(1.5),
  },
  buttonHalf: {
    flex: 1,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: "600",
    color: theme.color.textMuted,
    paddingHorizontal: theme.spacing(2),
    marginBottom: theme.spacing(1),
  },
  list: {
    paddingHorizontal: theme.spacing(2),
  },
  row: {
    paddingVertical: theme.spacing(1.5),
    borderBottomWidth: 1,
    borderBottomColor: theme.color.border,
  },
  rowText: {
    fontSize: 16,
    color: theme.color.text,
  },
  rowSub: {
    fontSize: 13,
    color: theme.color.textMuted,
    marginTop: 2,
  },
  empty: {
    color: theme.color.textMuted,
    paddingHorizontal: theme.spacing(2),
  },
});

export default InvoicesScreen;
