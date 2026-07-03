import { Routes, Route, Navigate } from "react-router-dom";
import "./App.css";
import Login from "./Pages/Login";
import SignUp from "./Pages/SignUp";
import { ToastContainer } from "react-toastify";
import Slidebar from "./Components/Dashboard/dash/Slidebar";

function App() {
  return (
    <>
      <div className="App">
        <Routes>
          <Route path="/" element={<Navigate to="/login" replace />} />
          <Route path="/login" element={<Login />} />
          <Route path="/signup" element={<SignUp />} />
          <Route path="/dashboard/*" element={<Slidebar />}>
            <Route path=":topics" element={<Slidebar />}></Route>
          </Route>
        </Routes>
      </div>
      <ToastContainer position="top-right" />
    </>
  );
}

export default App;
