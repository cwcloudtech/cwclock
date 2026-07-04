import React from "react";
import AuthNav from "../common/AuthNav";

const LoginNav = () => (
  <AuthNav prompt="Doesn't have an account?" linkTo="/signup" linkLabel="Sign Up" />
);

export default LoginNav;
