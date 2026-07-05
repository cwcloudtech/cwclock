import React from 'react';
import styles from './Styles/Headings.module.css';


const Heading = () => {
  return (
    <div className={styles.Section}>
       <h1 className={styles.title}>Get started with CWClock</h1>
       <p className={styles.subtitle}>Create an account to start tracking time and improve your productivity.</p>
    </div>
  )
}

export default Heading