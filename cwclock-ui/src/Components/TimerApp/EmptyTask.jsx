import React from 'react';
import styles from './Styles/TaskApp.module.css';
import logo from '../../assets/images/octopus.png';
import { useI18n } from '../../i18n/I18nContext';

const EmptyTask = () => {
  const { t } = useI18n();
  return (
    <div className={styles.Empty}>
        <img src={logo} alt="" />
        <h4>{t('timeTracker.emptyTitle')}</h4>
        <p>{t('timeTracker.emptyBody')}</p>
    </div>
  )
}

export default EmptyTask
