import React, { useEffect, useState } from "react";
import styles from "../Login/Styles/Form.module.css";
import Form from "react-bootstrap/Form";
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
    <div>
      <div className={styles.form}>
        <h5>Sign Up</h5>
        <Form onSubmit={handleSubmit}>
          {/* //email Input */}
          <Form.Group className="mb-3" controlId="formBasicEmail">
            <Form.Control
              onChange={onChange}
              name="email"
              value={email}
              type="email"
              placeholder="Enter email"
            />
          </Form.Group>

          <Form.Group className="mb-3" controlId="formBasicPassword">
            <Form.Control
              onChange={onChange}
              name="password"
              value={password}
              type="password"
              placeholder="Password"
            />
          </Form.Group>

          <Form.Group className="mb-3" controlId="formBasicConfirmPassword">
            <Form.Control
              onChange={onChange}
              name="confirmPassword"
              value={confirmPassword}
              type="password"
              placeholder="Confirm password"
            />
          </Form.Group>

          {error && <p className="cw-error">{error}</p>}

          <button type="submit" className={styles.btn}>
            CREATE ACCOUNT
          </button>
        </Form>
      </div>
    </div>
  );
};

export default SignUpForm;
