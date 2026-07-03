import React from 'react'
import Heading from '../Components/SignUp/Heading';
import SignUpForm from '../Components/SignUp/SignUpForm';
import SignUpNav from '../Components/SignUp/SignUpNav';
import Footer from '../Components/Login/Footer';

const SignUp = () => {
  return (
    <div style={{textAlign:"center"}}>
        <SignUpNav/>
        <Heading/>
        <SignUpForm/>
        <Footer/>
    </div>
  )
}

export default SignUp;