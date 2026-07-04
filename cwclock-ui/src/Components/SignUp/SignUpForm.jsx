import React, { useEffect, useState } from "react";
import styles from "../Login/Styles/Form.module.css";
import { useDispatch, useSelector } from "react-redux";
import { useNavigate } from "react-router-dom";
import Spinner from "../spinner/Spinner";
import { registerApi } from "../../Redux/Auth/Auth.actions";

const SignUpForm = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user, isLoading } = useSelector((state) => state.auth);
  const [formData, setFormData] = useState({
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [error, setError] = useState("");

  const { email, password, confirmPassword } = formData;

  const onChange = (e) => {
    let { name, value } = e.target;
    setFormData({
      ...formData,
      [name]: value,
    });
  };
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError("Passwords do not match.");
      return;
    }
    setError("");
    const userData = {
      email,
      password,
    };
    dispatch(registerApi(userData));
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
      <h1 className={styles.heading}>Sign Up</h1>
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

        <div className={styles.field}>
          <input
            className={styles.input}
            onChange={onChange}
            name="confirmPassword"
            value={confirmPassword}
            type="password"
            placeholder="Confirm password"
          />
        </div>

        {error && <p className="cw-error">{error}</p>}

        <button type="submit" className={styles.btn}>
          Create account
        </button>
      </form>
    </div>
  );
};

export default SignUpForm;
