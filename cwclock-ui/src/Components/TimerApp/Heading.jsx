import React from "react";
import styles from "./Styles/TaskComp.module.css";
import { useI18n } from "../../i18n/I18nContext";

const Heading = () => {
  const { t } = useI18n();
  const now = new Date();
  const h = now.getHours();
  const m = now.getMinutes();
  const s = now.getSeconds();
  const day = t(`days.${now.getDay()}`);

  return (
    <div className={styles.TaskHead}>
      <h6>{day}</h6>
      <div className={styles.Edit}>
        <h6 title={t("timeTracker.currentTime")}>
          {h < 10 ? "0" + h : h}:{m < 10 ? "0" + m : m}:{s < 10 ? "0" + s : s}
        </h6>
      </div>
    </div>
  );
};

export default Heading;
