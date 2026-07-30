import React, { useEffect, useMemo, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Screen from "../../components/Screen";
import FormField from "../../components/FormField";
import SelectField from "../../components/SelectField";
import Button from "../../components/Button";
import ErrorBanner from "../../components/ErrorBanner";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import { isOrgOwner } from "../../common/permissions";
import { listCountriesApi } from "../../redux/countries/countries.actions";
import { listCurrenciesApi } from "../../redux/currencies/currencies.actions";
import { updateOrganizationApi } from "../../redux/organizations/organizations.actions";

// OrganizationScreen: view/edit the current organization's profile. Ported
// from cwclock-ui's Organizations.jsx edit form, trimmed to the core fields
// (no logo/stamp upload - no image-picker dependency in this app - and no
// external-connections management, both left to the web app). Editing is
// owner-only, matching OrganizationHandler.Update's route gate
// (RequireRole(RoleOwner)) - other roles see a read-only view.
const OrganizationScreen = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { orgId, user } = useSelector((state) => state.session);
  const { items: orgs } = useSelector((state) => state.organizations);
  const { items: countries } = useSelector((state) => state.countries);
  const { items: currencies } = useSelector((state) => state.currencies);

  const org = orgs.find((o) => o.id === orgId);
  const canEdit = isOrgOwner(user, org);

  const [fields, setFields] = useState(null);
  const [error, setError] = useState(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    dispatch(listCountriesApi());
    dispatch(listCurrenciesApi());
  }, [dispatch]);

  useEffect(() => {
    if (org && !fields) {
      setFields({
        name: org.name || "",
        accountingEmail: org.accountingEmail || "",
        address: org.address || "",
        postalCode: org.postalCode || "",
        city: org.city || "",
        country: org.country || "",
        currency: org.currency || "",
        vatNumber: org.vatNumber || "",
        identificationNumber: org.identificationNumber || "",
        iban: org.iban || "",
        bic: org.bic || "",
      });
    }
  }, [org, fields]);

  const countryItems = useMemo(() => countries.map((c) => ({ value: c.iso, label: c.name })), [countries]);
  const currencyItems = useMemo(() => currencies.map((c) => ({ value: c.iso, label: c.name })), [currencies]);

  if (!org || !fields) return null;

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSave = async () => {
    setError(null);
    setSaving(true);
    try {
      await dispatch(updateOrganizationApi(orgId, fields));
    } catch (e) {
      setError(apiErrorMessage(e, locale));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Screen>
      {!canEdit && <ErrorBanner message={t("organizations.ownerOnlyHint")} />}
      <FormField
        label={t("organizations.name")}
        value={fields.name}
        onChangeText={(v) => setField("name", v)}
        editable={canEdit}
      />
      <FormField
        label={t("organizations.accountingEmail")}
        value={fields.accountingEmail}
        onChangeText={(v) => setField("accountingEmail", v)}
        autoCapitalize="none"
        keyboardType="email-address"
        editable={canEdit}
      />
      <FormField
        label={t("organizations.address")}
        value={fields.address}
        onChangeText={(v) => setField("address", v)}
        editable={canEdit}
      />
      <FormField
        label={t("organizations.postalCode")}
        value={fields.postalCode}
        onChangeText={(v) => setField("postalCode", v)}
        editable={canEdit}
      />
      <FormField
        label={t("organizations.city")}
        value={fields.city}
        onChangeText={(v) => setField("city", v)}
        editable={canEdit}
      />
      <SelectField
        label={t("organizations.country")}
        value={fields.country}
        onValueChange={(v) => setField("country", v)}
        items={countryItems}
        placeholder={t("organizations.country")}
        enabled={canEdit}
      />
      <SelectField
        label={t("organizations.currency")}
        value={fields.currency}
        onValueChange={(v) => setField("currency", v)}
        items={currencyItems}
        placeholder={t("organizations.currency")}
        enabled={canEdit}
      />
      <FormField
        label={t("organizations.vatNumber")}
        value={fields.vatNumber}
        onChangeText={(v) => setField("vatNumber", v)}
        editable={canEdit}
      />
      <FormField
        label={t("organizations.identificationNumber")}
        value={fields.identificationNumber}
        onChangeText={(v) => setField("identificationNumber", v)}
        editable={canEdit}
      />
      <FormField
        label={t("organizations.iban")}
        value={fields.iban}
        onChangeText={(v) => setField("iban", v)}
        autoCapitalize="characters"
        editable={canEdit}
      />
      <FormField
        label={t("organizations.bic")}
        value={fields.bic}
        onChangeText={(v) => setField("bic", v)}
        autoCapitalize="characters"
        editable={canEdit}
      />
      <ErrorBanner message={error} />
      {canEdit && <Button title={t("common.save")} onPress={handleSave} loading={saving} />}
    </Screen>
  );
};

export default OrganizationScreen;
