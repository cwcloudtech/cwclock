import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import Form from "react-bootstrap/Form";
import Button from "react-bootstrap/Button";
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
import { searchUsersApi } from "../../Redux/Users/User.actions";
import RequiredMark from "../common/RequiredMark";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";

const useEmailAutocomplete = (email, enabled, token) => {
  const dispatch = useDispatch();
  const [suggestions, setSuggestions] = useState([]);

  useEffect(() => {
    if (!enabled || email.length < 2) {
      setSuggestions([]);
      return;
    }
    const timeout = setTimeout(async () => {
      const results = await dispatch(searchUsersApi(email, token));
      setSuggestions(results || []);
    }, 300);
    return () => clearTimeout(timeout);
  }, [email, enabled, token, dispatch]);

  return suggestions;
};

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
};

const orgFormConfig = {
  name: "Organization",
  fields: [
    { name: "name", type: "text", label: "Name", required: true },
    { name: "address", type: "text", label: "Address" },
    { name: "postalCode", type: "text", label: "Postal code" },
    { name: "city", type: "text", label: "City" },
    { name: "country", type: "text", label: "Country" },
    { name: "vatNumber", type: "text", label: "VAT number" },
    { name: "siren", type: "text", label: "SIREN" },
    { name: "siret", type: "text", label: "SIRET" },
    { name: "picture", type: "image", label: "Picture" },
  ],
};

const MemberRow = ({ member, canSetRate, orgId, token }) => {
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
      <div>{member.email} - {member.role}</div>
      {canSetRate && (
        editing ? (
          <Form onSubmit={handleSave}>
            <Form.Group className="mb-2">
              <Form.Label>Daily rate</Form.Label>
              <Form.Control
                type="number"
                min="0"
                step="0.01"
                value={dailyRate}
                onChange={(e) => setDailyRate(e.target.value)}
              />
            </Form.Group>
            <Form.Group className="mb-2">
              <Form.Label>Currency</Form.Label>
              <Form.Control
                type="text"
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
              />
            </Form.Group>
            <Button size="sm" type="submit" className="me-2">
              Save
            </Button>
            <Button size="sm" variant="secondary" onClick={() => setEditing(false)}>
              Cancel
            </Button>
          </Form>
        ) : (
          <span>
            {member.dailyRate ? `${member.dailyRate} ${member.currency}/day` : "No daily rate set"}{" "}
            <Button size="sm" variant="link" onClick={() => setEditing(true)}>
              Edit
            </Button>
          </span>
        )
      )}
    </li>
  );
};

const Organizations = () => {
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

  useEffect(() => {
    dispatch(listOrgsApi(user.token));
  }, [dispatch, user.token]);

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
      setFields(emptyFields);
    } catch (err) {
      setCreateError("Name is required.");
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
      setEditError("Name is required.");
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
      setMemberError("Could not add member. Check the email and try again.");
    }
  };

  const handleTransferOwnership = async (e) => {
    e.preventDefault();
    setTransferError("");
    if (!transferEmail) return;
    if (!window.confirm(`Transfer ownership of "${currentOrg.name}" to ${transferEmail}?`)) {
      return;
    }
    try {
      await dispatch(transferOwnershipApi(currentOrgId, transferEmail, user.token));
      setTransferEmail("");
    } catch (err) {
      setTransferError("Could not transfer ownership. Check the email and try again.");
    }
  };

  return (
    <div className={styles.main}>
      <h1 className="cw-title">Organizations</h1>

      <ul className="cw-list">
        {organizations.map((org) => (
          <li key={org.id} className="cw-list-item">
            <label>
              <input
                type="radio"
                name="currentOrg"
                checked={org.id === currentOrgId}
                onChange={() => dispatch(selectOrg(org.id))}
              />{" "}
              {org.picture && <img src={org.picture} alt="" className={styles.avatar} />} {org.name}
            </label>
          </li>
        ))}
      </ul>

      <CollapsiblePanel title="Create an organization" defaultOpen={organizations.length === 0}>
        <ConfigForm
          config={orgFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleCreate}
          submitLabel="Create"
          error={createError}
        />
      </CollapsiblePanel>

      {currentOrg && (
        <div className={styles.members}>
          {isAdminOrOwner && (
            <>
              {editFields ? (
                <div>
                  <h2 className="cw-subtitle">Edit {currentOrg.name}</h2>
                  <ConfigForm
                    config={orgFormConfig}
                    values={editFields}
                    onChange={setEditField}
                    onSubmit={handleEditSubmit}
                    submitLabel="Save"
                    error={editError}
                  />
                  <Button variant="secondary" onClick={() => setEditFields(null)}>
                    Cancel
                  </Button>
                </div>
              ) : (
                <Button variant="outline-secondary" onClick={startEdit} className="mb-3">
                  Edit organization
                </Button>
              )}
            </>
          )}

          <h2 className="cw-subtitle">Members of {currentOrg.name}</h2>
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
              <CollapsiblePanel title="Add a member">
                <Form onSubmit={handleAddMember}>
                  <Form.Group className="mb-2">
                    <Form.Label>
                      Member email
                      <RequiredMark />
                    </Form.Label>
                    <Form.Control
                      list="member-email-suggestions"
                      type="email"
                      value={memberEmail}
                      onChange={(e) => setMemberEmail(e.target.value)}
                      required
                    />
                    <datalist id="member-email-suggestions">
                      {memberSuggestions.map((u) => (
                        <option key={u.id} value={u.email} />
                      ))}
                    </datalist>
                  </Form.Group>
                  <Form.Group className="mb-2">
                    <Form.Label>Role</Form.Label>
                    <Form.Select value={memberRole} onChange={(e) => setMemberRole(e.target.value)}>
                      <option value="admin">Admin</option>
                      <option value="member">Member</option>
                      <option value="reader">Reader</option>
                    </Form.Select>
                  </Form.Group>
                  <Button type="submit">Add member</Button>
                  {memberError && <p className="cw-error">{memberError}</p>}
                </Form>
              </CollapsiblePanel>

              <CollapsiblePanel title="Transfer ownership">
                <Form onSubmit={handleTransferOwnership}>
                  <Form.Group className="mb-2">
                    <Form.Label>
                      New owner's email
                      <RequiredMark />
                    </Form.Label>
                    <Form.Control
                      list="transfer-email-suggestions"
                      type="email"
                      value={transferEmail}
                      onChange={(e) => setTransferEmail(e.target.value)}
                      required
                    />
                    <datalist id="transfer-email-suggestions">
                      {transferSuggestions.map((u) => (
                        <option key={u.id} value={u.email} />
                      ))}
                    </datalist>
                  </Form.Group>
                  <Button type="submit" variant="warning">
                    Transfer ownership
                  </Button>
                  {transferError && <p className="cw-error">{transferError}</p>}
                </Form>
              </CollapsiblePanel>
            </>
          )}
        </div>
      )}
    </div>
  );
};

export default Organizations;
