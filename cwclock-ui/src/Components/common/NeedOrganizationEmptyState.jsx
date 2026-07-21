import React from "react";
import { useNavigate } from "react-router-dom";
import EmptyState from "./EmptyState";
import Button from "./Button";
import { useI18n } from "../../i18n/I18nContext";

// NeedOrganizationEmptyState is the octopus panel shown in place of a screen
// that requires an organization (time tracker, reports, invoices, projects,
// clients, ...) when none is selected yet (see ai-instruct-71) - the same
// EmptyState design used for the empty organizations list, with a button
// straight to where one gets created.
const NeedOrganizationEmptyState = ({ body }) => {
  const { t } = useI18n();
  const navigate = useNavigate();

  return (
    <EmptyState
      body={body}
      action={
        <Button onClick={() => navigate("/dashboard/organizations")} title={t("common.createOrganization")}>
          {t("common.createOrganization")}
        </Button>
      }
    />
  );
};

export default NeedOrganizationEmptyState;
