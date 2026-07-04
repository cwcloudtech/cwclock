import React from "react";
import AuthNav from "../common/AuthNav";

const LoginNav = () => (
  <AuthNav prompt="Don't have an account?" linkTo="/signup" linkLabel="Sign Up" />
);

export default LoginNav;
