import React, { useState, useEffect } from "react";
import styles from "./STYLE/Slidebar.module.css";
import { FaChevronLeft, FaChevronRight, FaUserCheck } from "react-icons/fa";
import { Route, Routes, useNavigate, Link } from "react-router-dom";
import Dropdown, { DropdownItem, DropdownText, DropdownDivider } from "../../common/Dropdown";
import logo from "../../../assets/images/cwclock-logo.svg";
import TimeTracker from "../pages/TimeTracker";
import SidebarNav from "./SidebarNav";

import Clientdiv from "../pages/Client";
import Organizationsdiv from "../pages/Organizations";
import Projectsdiv from "../pages/Projects";
import { useSelector, useDispatch } from "react-redux";
import { meApi, logoutUser } from "../../../Redux/Auth/Auth.actions";
import { updatePictureApi } from "../../../Redux/Users/User.actions";
import { listOrgsApi, selectOrg } from "../../../Redux/Organizations/Org.actions";
import fileToBase64 from "../../common/fileToBase64";

const Slidebar = () => {
  const [expanded, setExpanded] = useState(false);
  const handleclick = () => {
    setExpanded(!expanded);
  };

  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations, currentOrgId } = useSelector((state) => state.organizations);
  const currentOrg = organizations.find((o) => o.id === currentOrgId);

  useEffect(() => {
    if (!user.token) {
      navigate("/login");
      return;
    }
    dispatch(meApi(user.token));
    dispatch(listOrgsApi(user.token));
  }, [user, dispatch, navigate]);

  const handleLogout = () => {
    dispatch(logoutUser());
    navigate("/login");
  };

  const handlePictureChange = async (e) => {
    const file = e.target.files[0];
    if (!file) return;
    const picture = await fileToBase64(file);
    dispatch(updatePictureApi(picture, user.token));
  };

  return (
    <div className={styles.main}>
      <div className={styles.navbar}>
        <div className={styles.navbarleftmain}>
          <img src={logo} alt="cwclock logo" className={styles.logo} />

          {organizations.length > 0 ? (
            <Dropdown
              trigger={
                <>
                  {currentOrg?.picture && (
                    <img src={currentOrg.picture} alt="" className={styles.orgAvatar} />
                  )}
                  {currentOrg?.name}
                </>
              }
            >
              {(close) =>
                organizations.map((org) => (
                  <DropdownItem
                    key={org.id}
                    active={org.id === currentOrgId}
                    onClick={() => {
                      dispatch(selectOrg(org.id));
                      close();
                    }}
                  >
                    {org.picture && <img src={org.picture} alt="" className={styles.orgAvatar} />}
                    {org.name}
                  </DropdownItem>
                ))
              }
            </Dropdown>
          ) : (
            <Link className={styles.createOrgLink} to="/dashboard/organizations">
              Create organization
            </Link>
          )}
        </div>

        <div className={styles.navbarrightmain}>
          {user.token ? (
            <Dropdown
              align="end"
              triggerClassName={styles.profileTrigger}
              trigger={
                user.picture ? (
                  <img src={user.picture} alt="" className={styles.profileAvatar} />
                ) : (
                  <FaUserCheck fontSize="18px" />
                )
              }
            >
              <DropdownText>{user.email}</DropdownText>
              <DropdownText>
                <label className={styles.uploadLabel}>
                  Change picture
                  <input type="file" accept="image/*" hidden onChange={handlePictureChange} />
                </label>
              </DropdownText>
              <DropdownDivider />
              <DropdownItem onClick={handleLogout}>Logout</DropdownItem>
            </Dropdown>
          ) : (
            <button className={styles.loginBtn} onClick={() => navigate("/login")}>
              Login
            </button>
          )}
        </div>
      </div>

      <div className={styles.Slideflex}>
        <div className={`${styles.sidebarCol} ${expanded ? styles.sidebarColExpanded : ""}`}>
          <SidebarNav expanded={expanded} />
          <button
            className={styles.sidebarToggle}
            onClick={handleclick}
            title={expanded ? "Collapse sidebar" : "Expand sidebar"}
          >
            {expanded ? <FaChevronLeft /> : <FaChevronRight />}
          </button>
        </div>
        <div className={styles.pages}>
          <Routes>
            <Route path="/timetracker" element={<TimeTracker />}></Route>
            <Route path="/organizations" element={<Organizationsdiv />}></Route>
            <Route path="/clients" element={<Clientdiv />}></Route>
            <Route path="/projects" element={<Projectsdiv />}></Route>
          </Routes>
        </div>
      </div>
    </div>
  );
};

export default Slidebar;
