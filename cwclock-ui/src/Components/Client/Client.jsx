import React, { useEffect, useState } from "react";
import { useDispatch, useSelector } from "react-redux";
import styles from "./Styles/Client.module.css";
import { listClientsApi, createClientApi } from "../../Redux/Clients/Client.actions";
import ConfigForm from "../common/ConfigForm";
import CollapsiblePanel from "../common/CollapsiblePanel";

const initialFields = {
  name: "",
  address: "",
  postalCode: "",
  city: "",
  country: "",
  vatNumber: "",
  vatRate: "",
  vatDischargeMotive: "",
  purchaseOrder: "",
  hoursPerDay: "",
};

const clientFormConfig = {
  name: "Client",
  fields: [
    { name: "name", type: "text", label: "Name", required: true },
    { name: "address", type: "text", label: "Address" },
    { name: "postalCode", type: "text", label: "Postal code" },
    { name: "city", type: "text", label: "City" },
    { name: "country", type: "text", label: "Country" },
    { name: "vatNumber", type: "text", label: "VAT number" },
    { name: "vatRate", type: "number", label: "VAT rate % (default 20)", step: "0.01" },
    { name: "vatDischargeMotive", type: "text", label: "VAT discharge motive" },
    { name: "purchaseOrder", type: "text", label: "Purchase order" },
    { name: "hoursPerDay", type: "number", label: "Hours/day (default 7)", step: "0.01" },
  ],
};

const Clients = () => {
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId } = useSelector((state) => state.organizations);
  const { clients } = useSelector((state) => state.clients);
  const [fields, setFields] = useState(initialFields);
  const [error, setError] = useState("");

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listClientsApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

  const setField = (key, value) => setFields({ ...fields, [key]: value });

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError("");
    if (!currentOrgId) return;
    const payload = {
      ...fields,
      vatRate: fields.vatRate === "" ? undefined : Number(fields.vatRate),
      hoursPerDay: fields.hoursPerDay === "" ? undefined : Number(fields.hoursPerDay),
    };
    try {
      await dispatch(createClientApi(currentOrgId, payload, user.token));
      setFields(initialFields);
    } catch (err) {
      setError("Name is required.");
    }
  };

  if (!currentOrgId) {
    return <h1 className="cw-title">Select or create an organization first</h1>;
  }

  return (
    <div className={styles.main}>
      <h1 className="cw-title">Clients</h1>
      <CollapsiblePanel title="Create a client">
        <ConfigForm
          config={clientFormConfig}
          values={fields}
          onChange={setField}
          onSubmit={handleSubmit}
          submitLabel="ADD"
          error={error}
        />
      </CollapsiblePanel>

      <ul className="cw-list">
        {clients.map((client) => (
          <li className="cw-list-item" key={client.id}>
            <strong>{client.name}</strong>
            {client.address && `, ${client.address}`}
            {client.postalCode && ` ${client.postalCode}`}
            {client.city && ` ${client.city}`}
            {client.country && ` ${client.country}`}
            {" - VAT "}{client.vatRate}%{" - "}{client.hoursPerDay}h/day
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Clients;
