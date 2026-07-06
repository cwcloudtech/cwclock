import React from "react";
import AuthNav from "../common/AuthNav";
import { useI18n } from "../../i18n/I18nContext";

const SignUpNav = () => {
  const { t } = useI18n();
  return <AuthNav prompt={t("auth.alreadyHaveAccount")} linkTo="/login" linkLabel={t("auth.logIn")} />;
};

export default SignUpNav;
