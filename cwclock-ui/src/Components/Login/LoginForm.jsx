import React, { useEffect, useState } from "react";
import styles from "./Styles/Form.module.css";
import { useNavigate } from "react-router-dom";
import { useDispatch, useSelector } from "react-redux";
import Spinner from "../spinner/Spinner";
import { loginApi } from "../../Redux/Auth/Auth.actions";

const LoginForm = ({ label, checkBox }) => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isLoading } = useSelector((state) => state.auth);
  const [formData, setFormData] = useState({
    email: "",
    password: "",
  });
  const { email, password } = formData;

  const onChange = (e) => {
    let { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };
  const handleSubmit = (e) => {
    e.preventDefault();
    const userData = {
      email,
      password,
    };
    dispatch(loginApi(userData));
  };
  useEffect(() => {
    if (user.token) {
      navigate("/dashboard/timetracker");
    }
  }, [user, navigate]);

  if (isLoading) {
    return <Spinner />;
  }
  return (
    <div className={styles.form}>
      <h1 className={styles.heading}>{label}</h1>
      <form onSubmit={handleSubmit}>
        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="email"
            value={email}
            type="email"
            placeholder="Enter email"
          />
        </div>

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="password"
            value={password}
            type="password"
            placeholder="Password"
          />
        </div>

        <label className={styles.formflex}>
          <input type="checkbox" />
          {checkBox}
        </label>

        <button type="submit" className={styles.btn}>
          Log In
        </button>
      </form>
    </div>
  );
};

export default LoginForm;
