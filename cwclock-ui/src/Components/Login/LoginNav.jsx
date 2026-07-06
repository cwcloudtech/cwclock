import React from "react";
import AuthNav from "../common/AuthNav";
import { useI18n } from "../../i18n/I18nContext";

const LoginNav = () => {
  const { t } = useI18n();
  return <AuthNav prompt={t("auth.doesNotHaveAccount")} linkTo="/signup" linkLabel={t("auth.signUp")} />;
};

export default LoginNav;
