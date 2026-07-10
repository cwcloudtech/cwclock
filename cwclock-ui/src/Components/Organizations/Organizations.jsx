import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Button from "../common/Button";
import ConfirmModal from "../common/ConfirmModal";
import CopyIdButton from "../common/CopyIdButton";
import styles from "./Styles/Organizations.module.css";
import {
  listOrgsApi,
  createOrgApi,
  updateOrgApi,
  selectOrg,
  listMembersApi,
  addMemberApi,
  transferOwnershipApi,
  setMemberRateApi,
} from "../../Redux/Organizations/Org.actions";
import RequiredMark from "../common/RequiredMark";
import ConfigForm from "../common/ConfigForm";
import applyImageField from "../common/applyImageField";
import { isAdminOrOwner as computeIsAdminOrOwner, isOrgOwner } from "../common/permissions";
import CollapsiblePanel from "../common/CollapsiblePanel";
import EmptyState from "../common/EmptyState";
import memberLabel from "../common/memberLabel";
import useEmailAutocomplete from "../common/useEmailAutocomplete";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import useCurrencies from "../common/useCurrencies";
import useCountries from "../common/useCountries";
import useCountryFields from "../common/useCountryFields";
import { identificationFieldConfig } from "../common/identificationFields";

const emptyFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  siren: "",
  siret: "",
  naf: "",
  mf: "",
  identificationNumber: "",
  picture: "",
  pictureX: 50,
  pictureY: 50,
  stamp: "",
  stampX: 50,
  stampY: 50,
  currency: "",
};

const roleLabelKey = {
  owner: "common.roleOwner",
  admin: "common.roleAdmin",
  member: "common.roleMember",
  reader: "common.roleReader",
};

const MemberRow = ({ member, canSetRate, orgId, orgCurrency, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [editing, setEditing] = useState(false);
  const [dailyRate, setDailyRate] = useState(member.dailyRate || "");

  const handleSave = (e) => {
    e.preventDefault();
    if (!dailyRate || dailyRate <= 0) return;
    dispatch(setMemberRateApi(orgId, member.userId, Number(dailyRate), token));
    setEditing(false);
  };

  return (
    <li className="cw-list-item">
      <div className={styles.memberRow}>
        <span>
          {memberLabel(member)}
          {member.name && ` (${member.email})`} - {t(roleLabelKey[member.role] || "common.roleMember")}
        </span>
        {canSetRate && !editing && (
          <span className={styles.rate}>
            {member.dailyRate
              ? t("organizations.dailyRatePerDay", { rate: member.dailyRate, currency: orgCurrency })
              : t("organizations.noDailyRateSet")}{" "}
            <Button size="sm" variant="ghost" onClick={() => setEditing(true)} title={t("organizations.editDailyRate")}>
              {t("common.edit")}
            </Button>
          </span>
        )}
      </div>
      {canSetRate && editing && (
        <form className={styles.rateForm} onSubmit={handleSave}>
          <div className="cw-field">
            <label className="cw-label">{t("organizations.dailyRate")}</label>
            <input
              className="cw-input"
              type="number"
              min="0"
              step="0.01"
              value={dailyRate}
              onChange={(e) => setDailyRate(e.target.value)}
              title={t("organizations.dailyRate")}
            />
          </div>
          <div className="cw-field">
            <Button size="sm" type="submit" title={t("organizations.saveDailyRate")}>
              {t("common.save")}
            </Button>
          </div>
          <div className="cw-field">
            <Button size="sm" variant="secondary" onClick={() => setEditing(false)} title={t("common.discardChanges")}>
              {t("common.cancel")}
            </Button>
          </div>
        </form>
      )}
    </li>
  );
};

const Organizations = () => {
  const { t, locale } = useI18n();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations, members, currentOrgId } = useSelector((state) => state.organizations);
  const [fields, setFields] = useState(emptyFields);
  const [createError, setCreateError] = useState("");
  const [editFields, setEditFields] = useState(null);
  const [editError, setEditError] = useState("");
  const [memberEmail, setMemberEmail] = useState("");
  const [memberRole, setMemberRole] = useState("member");
  const [memberError, setMemberError] = useState("");
  const [transferEmail, setTransferEmail] = useState("");
  const [transferError, setTransferError] = useState("");
  const [confirmingTransfer, setConfirmingTransfer] = useState(false);
  const currencies = useCurrencies();
  const countries = useCountries();
  const createIdentificationFields = useCountryFields(fields.country);
  const editIdentificationFields = useCountryFields(editFields?.country);

  const countryOptions = countries.map((c) => ({ value: c.iso, label: c.name }));
  const currencyOptions = currencies.map((c) => ({ value: c.iso, label: c.iso }));

  const buildOrgFormConfig = (identificationFields) => ({
    name: "Organization",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      {
        name: "country",
        type: "autocomplete",
        label: t("common.country"),
        placeholder: t("common.country"),
        required: true,
        options: countryOptions,
      },
      ...identificationFields.map((name) => identificationFieldConfig(name, t)),
      {
        name: "currency",
        type: "select",
        label: t("common.currency"),
        required: true,
        options: currencyOptions,
      },
      { name: "picture", type: "image", label: t("common.picture") },
      { name: "stamp", type: "image", label: t("organizations.stamp") },
    ],
  });

  const orgFormConfig = buildOrgFormConfig(createIdentificationFields);
  const editOrgFormConfig = buildOrgFormConfig(editIdentificationFields);

  useEffect(() => {
    dispatch(listOrgsApi(user.token));
  }, [dispatch, user.token]);

  // Defaults the currency from the selected country (ai-instruct-35: "the
  // default currency should be selected according to the country but let
  // the user decide"), only while the field is still blank so it never
  // clobbers a currency the user picked themselves.
  useEffect(() => {
    if (!fields.country || fields.currency) return;
    const country = countries.find((c) => c.iso === fields.country);
    if (country) setFields((f) => ({ ...f, currency: country.currency }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [fields.country, countries]);

  useEffect(() => {
    if (!editFields || !editFields.country || editFields.currency) return;
    const country = countries.find((c) => c.iso === editFields.country);
    if (country) setEditFields((f) => ({ ...f, currency: country.currency }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [editFields?.country, countries]);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const currentOrg = organizations.find((o) => o.id === currentOrgId);
  const isOwner = isOrgOwner(user, currentOrg);
  const isAdminOrOwner = computeIsAdminOrOwner(user, members);

  const memberSuggestions = useEmailAutocomplete(memberEmail, isOwner, user.token);
  const transferSuggestions = useEmailAutocomplete(transferEmail, isOwner, user.token);

  const setField = (key, value) => setFields((f) => applyImageField(f, key, value));
  const setEditField = (key, value) => setEditFields((f) => applyImageField(f, key, value));

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreateError("");
    try {
      await dispatch(createOrgApi(fields, user.token));
      setFields(emptyFields);
    } catch (err) {
      setCreateError(apiErrorMessage(err, locale));
    }
  };

  const startEdit = () => {
    setEditError("");
    setEditFields({
      name: currentOrg.name || "",
      address: currentOrg.address || "",
      postalCode: currentOrg.postalCode || "",
      city: currentOrg.city || "",
      country: currentOrg.country || "",
      vatNumber: currentOrg.vatNumber || "",
      siren: currentOrg.siren || "",
      siret: currentOrg.siret || "",
      naf: currentOrg.naf || "",
      mf: currentOrg.mf || "",
      identificationNumber: currentOrg.identificationNumber || "",
      currency: currentOrg.currency || "",
      picture: currentOrg.picture || "",
      pictureX: currentOrg.pictureX ?? 50,
      pictureY: currentOrg.pictureY ?? 50,
      stamp: currentOrg.stamp || "",
      stampX: currentOrg.stampX ?? 50,
      stampY: currentOrg.stampY ?? 50,
    });
  };

  const handleEditSubmit = async (e) => {
    e.preventDefault();
    setEditError("");
    try {
      await dispatch(updateOrgApi(currentOrgId, editFields, user.token));
      setEditFields(null);
    } catch (err) {
      setEditError(apiErrorMessage(err, locale));
    }
  };

  const handleAddMember = async (e) => {
    e.preventDefault();
    setMemberError("");
    if (!memberEmail) return;
    try {
      await dispatch(addMemberApi(currentOrgId, memberEmail, memberRole, user.token));
      setMemberEmail("");
    } catch (err) {
      setMemberError(apiErrorMessage(err, locale));
    }
  };

  const handleTransferClick = (e) => {
    e.preventDefault();
    setTransferError("");
    if (!transferEmail) return;
    setConfirmingTransfer(true);
  };

  const handleTransferConfirm = async () => {
    setConfirmingTransfer(false);
    try {
      await dispatch(transferOwnershipApi(currentOrgId, transferEmail, user.token));
      setTransferEmail("");
    } catch (err) {
      setTransferError(apiErrorMessage(err, locale));
    }
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">{t("organizations.title")}</h1>

      {organizations.length === 0 && (
        <EmptyState title={t("organizations.emptyTitle")} body={t("organizations.emptyBody")} />
      )}

      <ul className="cw-list">
        {organizations.map((org) => (
          <li key={org.id} className={`cw-list-item ${styles.orgRow}`}>
            <label className={styles.orgOption} title={t("organizations.switchTo", { name: org.name })}>
              <input
                type="radio"
                name="currentOrg"
                checked={org.id === currentOrgId}
                onChange={() => dispatch(selectOrg(org.id))}
              />
              {org.picture && (
                <img
                  src={org.picture}
                  alt=""
                  className={styles.avatar}
                  style={{ objectPosition: `${org.pictureX ?? 50}% ${org.pictureY ?? 50}%` }}
                />
              )}
              {org.name}
            </label>
            <div className={styles.rowActions}>
              <CopyIdButton id={org.id} className={styles.iconBtn} />
            </div>
          </li>
        ))}
      </ul>

      <CollapsiblePanel title={t("organizations.createOrganization")} defaultOpen={organizations.length === 0}>
        <ConfigForm
          config={orgFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleCreate}
          submitLabel={t("common.create")}
          error={createError}
        />
      </CollapsiblePanel>

      {currentOrg && (
        <div className={styles.members}>
          {isAdminOrOwner && (
            <>
              {editFields ? (
                <div>
                  <h2 className="cw-subtitle">{t("organizations.editOrgTitle", { name: currentOrg.name })}</h2>
                  <ConfigForm
                    config={editOrgFormConfig}
                    values={editFields}
                    onChange={setEditField}
                    onSubmit={handleEditSubmit}
                    submitLabel={t("common.save")}
                    onCancel={() => setEditFields(null)}
                    error={editError}
                  />
                </div>
              ) : (
                <Button
                  variant="secondary"
                  onClick={startEdit}
                  className={styles.editOrgBtn}
                  title={t("organizations.editOrgTooltip")}
                >
                  {t("organizations.editOrganization")}
                </Button>
              )}
            </>
          )}

          <h2 className="cw-subtitle">{t("organizations.membersOf", { name: currentOrg.name })}</h2>
          <ul className="cw-list">
            {members.map((m) => (
              <MemberRow
                key={m.id}
                member={m}
                canSetRate={isAdminOrOwner}
                orgId={currentOrgId}
                orgCurrency={currentOrg.currency}
                token={user.token}
              />
            ))}
          </ul>

          {isOwner && (
            <>
              <CollapsiblePanel title={t("organizations.addMemberPanel")}>
                <form onSubmit={handleAddMember}>
                  <div className="cw-field">
                    <label className="cw-label">
                      {t("organizations.memberEmail")}
                      <RequiredMark />
                    </label>
                    <input
                      className="cw-input"
                      list="member-email-suggestions"
                      type="email"
                      value={memberEmail}
                      onChange={(e) => setMemberEmail(e.target.value)}
                      required
                      title={t("organizations.memberEmail")}
                    />
                    <datalist id="member-email-suggestions">
                      {memberSuggestions.map((u) => (
                        <option key={u.id} value={u.email} label={memberLabel(u)} />
                      ))}
                    </datalist>
                  </div>
                  <div className="cw-field">
                    <label className="cw-label">{t("common.role")}</label>
                    <select
                      className="cw-select"
                      value={memberRole}
                      onChange={(e) => setMemberRole(e.target.value)}
                      title={t("common.role")}
                    >
                      <option value="admin">{t("common.roleAdmin")}</option>
                      <option value="member">{t("common.roleMember")}</option>
                      <option value="reader">{t("common.roleReader")}</option>
                    </select>
                  </div>
                  <Button type="submit" title={t("organizations.inviteMember")}>
                    {t("organizations.addMemberButton")}
                  </Button>
                  {memberError && <p className="cw-error">{memberError}</p>}
                </form>
              </CollapsiblePanel>

              <CollapsiblePanel title={t("organizations.transferOwnership")}>
                <form onSubmit={handleTransferClick}>
                  <div className="cw-field">
                    <label className="cw-label">
                      {t("organizations.newOwnerEmail")}
                      <RequiredMark />
                    </label>
                    <input
                      className="cw-input"
                      list="transfer-email-suggestions"
                      type="email"
                      value={transferEmail}
                      onChange={(e) => setTransferEmail(e.target.value)}
                      required
                      title={t("organizations.newOwnerEmail")}
                    />
                    <datalist id="transfer-email-suggestions">
                      {transferSuggestions.map((u) => (
                        <option key={u.id} value={u.email} label={memberLabel(u)} />
                      ))}
                    </datalist>
                  </div>
                  <Button type="submit" variant="danger" title={t("organizations.transferOwnershipTooltip")}>
                    {t("organizations.transferOwnership")}
                  </Button>
                  {transferError && <p className="cw-error">{transferError}</p>}
                </form>
              </CollapsiblePanel>

              <ConfirmModal
                show={confirmingTransfer}
                title={t("organizations.transferOwnership")}
                body={t("organizations.confirmTransfer", { name: currentOrg.name, email: transferEmail })}
                confirmLabel={t("organizations.transferOwnership")}
                onConfirm={handleTransferConfirm}
                onCancel={() => setConfirmingTransfer(false)}
              />
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default Organizations;
