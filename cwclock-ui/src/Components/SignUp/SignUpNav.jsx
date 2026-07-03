import React from 'react';
import styles from '../Login/Styles/LoginNav.module.css';
import {Link} from 'react-router-dom';
import logo from '../../assets/images/cwclock-logo.svg';

const SignUpNav = () => {
  return (
    <div className={styles.navW}>
    <div className={styles.Nav}>
        <Link to="/login" className={styles.Logo}>
        <img src={logo} alt="" />
        </Link>
        <div className={styles.Navflex}>
            <h6>Already have an account?</h6>
            <Link to="/login" className={styles.signUp}>Log In</Link>
        </div>
    </div>
    </div>
  )
}

export default SignUpNav;