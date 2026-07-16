import { Routes, Route, Navigate } from "react-router-dom";
import "./App.css";
import Login from "./Pages/Login";
import SignUp from "./Pages/SignUp";
import OidcCallback from "./Pages/OidcCallback";
import { ToastContainer } from "react-toastify";
import Slidebar from "./Components/Dashboard/dash/Slidebar";
import { ThemeProvider } from "./Components/common/ThemeContext";
import { I18nProvider } from "./i18n/I18nContext";
import CookieBanner from "./Components/common/CookieBanner";

function App() {
  return (
    <I18nProvider>
      <ThemeProvider>
        <div className="App">
          <Routes>
            <Route path="/" element={<Navigate to="/login" replace />} />
            <Route path="/login" element={<Login />} />
            <Route path="/signup" element={<SignUp />} />
            <Route path="/oidc/callback" element={<OidcCallback />} />
            <Route path="/dashboard/*" element={<Slidebar />}>
              <Route path=":topics" element={<Slidebar />}></Route>
            </Route>
          </Routes>
        </div>
        <ToastContainer position="top-right" />
        <CookieBanner />
      </ThemeProvider>
    </I18nProvider>
  );
}

export default App;
