import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Button from "../common/Button";
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
import CollapsiblePanel from "../common/CollapsiblePanel";
import memberLabel from "../common/memberLabel";
import useEmailAutocomplete from "../common/useEmailAutocomplete";
import { useI18n } from "../../i18n/I18nContext";
import { apiErrorMessage } from "../../i18n/translate";
import useCurrencies from "../common/useCurrencies";

const emptyFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  siren: "",
  siret: "",
  picture: "",
  currency: "",
};

const roleLabelKey = {
  owner: "common.roleOwner",
  admin: "common.roleAdmin",
  member: "common.roleMember",
  reader: "common.roleReader",
};

const MemberRow = ({ member, canSetRate, orgId, token }) => {
  const { t } = useI18n();
  const dispatch = useDispatch();
  const [editing, setEditing] = useState(false);
  const [dailyRate, setDailyRate] = useState(member.dailyRate || "");
  const [currency, setCurrency] = useState(member.currency || "euros");

  const handleSave = (e) => {
    e.preventDefault();
    if (!dailyRate || dailyRate <= 0) return;
    dispatch(setMemberRateApi(orgId, member.userId, Number(dailyRate), currency, token));
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
              ? t("organizations.dailyRatePerDay", { rate: member.dailyRate, currency: member.currency })
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
            <label className="cw-label">{t("organizations.currency")}</label>
            <input
              className="cw-input"
              type="text"
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              title={t("organizations.currency")}
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
  const currencies = useCurrencies();

  const orgFormConfig = {
    name: "Organization",
    fields: [
      { name: "name", type: "text", label: t("common.name"), required: true },
      { name: "address", type: "text", label: t("common.address") },
      { name: "postalCode", type: "text", label: t("common.postalCode") },
      { name: "city", type: "text", label: t("common.city") },
      { name: "country", type: "text", label: t("common.country") },
      { name: "vatNumber", type: "text", label: t("common.vatNumber") },
      { name: "siren", type: "text", label: "SIREN" },
      { name: "siret", type: "text", label: "SIRET" },
      {
        name: "currency",
        type: "select",
        label: t("common.currency"),
        required: true,
        options: currencies.map((code) => ({ value: code, label: code })),
      },
      { name: "picture", type: "image", label: t("common.picture") },
    ],
  };

  useEffect(() => {
    dispatch(listOrgsApi(user.token));
  }, [dispatch, user.token]);

  useEffect(() => {
    if (currencies.length && !fields.currency) {
      setFields((f) => ({ ...f, currency: currencies[0] }));
    }
    // Only react to the currency list becoming available, not to every
    // keystroke in the rest of the create-organization form.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currencies]);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const currentOrg = organizations.find((o) => o.id === currentOrgId);
  const myRole = members.find((m) => m.userId === user.id)?.role;
  const isOwner = currentOrg && currentOrg.ownerId === user.id;
  const isAdminOrOwner = isOwner || myRole === "admin";

  const memberSuggestions = useEmailAutocomplete(memberEmail, isOwner, user.token);
  const transferSuggestions = useEmailAutocomplete(transferEmail, isOwner, user.token);

  const setField = (key, value) => setFields({ ...fields, [key]: value });
  const setEditField = (key, value) => setEditFields({ ...editFields, [key]: value });

  const handleCreate = async (e) => {
    e.preventDefault();
    setCreateError("");
    try {
      await dispatch(createOrgApi(fields, user.token));
      setFields({ ...emptyFields, currency: currencies[0] || "" });
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
      currency: currentOrg.currency || currencies[0] || "",
      picture: currentOrg.picture || "",
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

  const handleTransferOwnership = async (e) => {
    e.preventDefault();
    setTransferError("");
    if (!transferEmail) return;
    if (!window.confirm(t("organizations.confirmTransfer", { name: currentOrg.name, email: transferEmail }))) {
      return;
    }
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

      <ul className="cw-list">
        {organizations.map((org) => (
          <li key={org.id} className="cw-list-item">
            <label className={styles.orgOption} title={t("organizations.switchTo", { name: org.name })}>
              <input
                type="radio"
                name="currentOrg"
                checked={org.id === currentOrgId}
                onChange={() => dispatch(selectOrg(org.id))}
              />
              {org.picture && <img src={org.picture} alt="" className={styles.avatar} />}
              {org.name}
            </label>
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
                    config={orgFormConfig}
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
                <form onSubmit={handleTransferOwnership}>
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
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default Organizations;
