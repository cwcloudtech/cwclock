import React, { useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import DateField from "../../components/DateField";
import SelectField from "../../components/SelectField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { toDayString } from "../../common/format";
import { generateReportPdfApi } from "../../redux/reports/reports.actions";

const firstOfMonth = () => {
  const d = new Date();
  d.setDate(1);
  return d;
};

// ReportsScreen: the "simplified generation screen for summary and detailed
// reports in PDF format only" from the ai-instruct-105 plan - just a report
// type and a date range, no client/project/user filters like the web
// Reports page has.
const ReportsScreen = ({ navigation }) => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId } = useSelector((state) => state.session);

  const [reportType, setReportType] = useState("summary");
  const [start, setStart] = useState(firstOfMonth);
  const [end, setEnd] = useState(() => new Date());
  const [error, setError] = useState(null);
  const [generating, setGenerating] = useState(false);

  const handleGenerate = async () => {
    setError(null);
    setGenerating(true);
    try {
      const path = await dispatch(
        generateReportPdfApi(orgId, reportType, { start: toDayString(start), end: toDayString(end) })
      );
      navigation.navigate("PdfViewer", { path, title: t(`reports.${reportType}`) });
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setGenerating(false);
    }
  };

  return (
    <Screen>
      <SelectField
        label={t("reports.type")}
        value={reportType}
        onValueChange={setReportType}
        items={[
          { value: "summary", label: t("reports.summary") },
          { value: "detailed", label: t("reports.detailed") },
        ]}
      />
      <DateField label={t("reports.startDate")} value={start} onChange={setStart} mode="date" />
      <DateField label={t("reports.endDate")} value={end} onChange={setEnd} mode="date" />
      <ErrorBanner message={error} />
      <Button title={t("reports.generate")} onPress={handleGenerate} loading={generating} />
    </Screen>
  );
};

export default ReportsScreen;
