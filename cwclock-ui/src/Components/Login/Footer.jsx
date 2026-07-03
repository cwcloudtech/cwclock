import React from 'react';
import styles from './Styles/Form.module.css';


const Footer = ({margin}) => {
  const REPO_URL = process.env.REACT_APP_REPOURL;

  return (
    <div>
    <div className={styles.footer} style={{marginTop:margin}}>
        <p>Open source solution, sources are available <a href={REPO_URL} target='_blank' rel='noreferrer'>here</a></p>
    </div>
    </div>
  )
}

export default Footer;