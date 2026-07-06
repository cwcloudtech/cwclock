import React from 'react';
import styles from './Styles/Headings.module.css';
import { useI18n } from '../../i18n/I18nContext';


const Heading = () => {
  const { t } = useI18n();
  return (
    <div className={styles.Section}>
       <h1 className={styles.title}>{t('auth.getStartedTitle')}</h1>
       <p className={styles.subtitle}>{t('auth.getStartedSubtitle')}</p>
    </div>
  )
}

export default Heading
