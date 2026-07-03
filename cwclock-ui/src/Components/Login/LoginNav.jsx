import React from 'react';
import styles from './Styles/LoginNav.module.css';
import {Link} from 'react-router-dom';
import logo from '../../assets/images/cwclock-logo.svg';

const LoginNav = () => {
  return (
    <div className={styles.Nav}>
        <Link to="/login" className={styles.Logo}>
        <img src={logo} alt="" />
        </Link>
        <div className={styles.Navflex}>
            <h6>Don't have an account?</h6>
            <Link to="/signup" className={styles.signUp}>Sign Up</Link>
        </div>
    </div>
  )
}

export default LoginNav;