import React from 'react';
import styles from './Styles/Form.module.css';
import { useI18n } from '../../i18n/I18nContext';


const Footer = ({margin}) => {
  const REPO_URL = process.env.REACT_APP_REPOURL;
  const { t } = useI18n();

  return (
    <div>
    <div className={styles.footer} style={{marginTop:margin}}>
        <p>{t('auth.openSourcePrefix')} <a href={REPO_URL} target='_blank' rel='noreferrer'>{t('auth.here')}</a></p>
    </div>
    </div>
  )
}

export default Footer;
