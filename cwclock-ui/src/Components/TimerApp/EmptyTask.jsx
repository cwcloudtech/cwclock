import React from 'react';
import styles from './Styles/TaskApp.module.css';
import logo from '../../assets/images/octopus.png';

const EmptyTask = () => {
  return (
    <div className={styles.Empty}>
        <img src={logo} alt="" />
        <h4>Let’s start tracking!</h4>
        <p>Pick a project above and start the timer.</p>
    </div>
  )
}

export default EmptyTask