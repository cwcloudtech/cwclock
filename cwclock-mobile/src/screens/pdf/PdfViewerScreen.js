import React, { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import Pdf from "react-native-pdf";
import Share from "react-native-share";
import Button from "../../components/Button";
import { useI18n } from "../../i18n/I18nContext";
import theme from "../../theme";

// PdfViewerScreen is the generic "show a locally-downloaded PDF" screen,
// reused by ReportsScreen, InvoicesScreen's Preview/Generate and its
// previous-invoices list - all of them just hand it a local file path (see
// src/api/blobRequest.js) and a title.
const PdfViewerScreen = ({ route }) => {
  const { t } = useI18n();
  const { path } = route.params;
  const [failed, setFailed] = useState(false);

  const handleShare = () => {
    const url = path.startsWith("file://") ? path : `file://${path}`;
    Share.open({ url, type: "application/pdf" }).catch(() => {});
  };

  return (
    <View style={styles.flex}>
      {failed ? (
        <View style={styles.center}>
          <Text style={styles.error}>{t("pdf.failedToLoad")}</Text>
        </View>
      ) : (
        <Pdf source={{ uri: path }} style={styles.flex} onError={() => setFailed(true)} />
      )}
      <View style={styles.actions}>
        <Button title={t("common.share")} onPress={handleShare} />
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  flex: { flex: 1 },
  center: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: theme.spacing(3),
  },
  error: {
    color: theme.color.danger,
    textAlign: "center",
  },
  actions: {
    padding: theme.spacing(2),
  },
});

export default PdfViewerScreen;
