import React, { useState, useEffect } from "react";
import styles from "./STYLE/Slidebar.module.css";
import { FaChevronLeft, FaChevronRight, FaUserCheck } from "react-icons/fa";
import { FiLogOut } from "react-icons/fi";
import { Route, Routes, useNavigate, Link } from "react-router-dom";
import Dropdown, { DropdownItem, DropdownText, DropdownDivider } from "../../common/Dropdown";
import EditProfileModal from "../../common/EditProfileModal";
import DisabledNotice from "../../common/DisabledNotice";
import memberLabel from "../../common/memberLabel";
import Tooltip from "../../common/Tooltip";
import logo from "../../../assets/images/cwclock-logo.svg";
import TimeTracker from "../pages/TimeTracker";
import SidebarNav from "./SidebarNav";

import Clientdiv from "../pages/Client";
import Organizationsdiv from "../pages/Organizations";
import Projectsdiv from "../pages/Projects";
import Admindiv from "../pages/Admin";
import AdminOrganizationsdiv from "../pages/AdminOrganizations";
import { useSelector, useDispatch } from "react-redux";
import { meApi, logoutUser } from "../../../Redux/Auth/Auth.actions";
import { listOrgsApi, selectOrg } from "../../../Redux/Organizations/Org.actions";

const Slidebar = () => {
  const [expanded, setExpanded] = useState(false);
  const [showEditProfile, setShowEditProfile] = useState(false);
  const handleclick = () => {
    setExpanded(!expanded);
  };

  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { user } = useSelector((state) => state.auth);
  const { organizations, currentOrgId } = useSelector((state) => state.organizations);
  const currentOrg = organizations.find((o) => o.id === currentOrgId);
  const isSuperuser = user.role === "superuser";

  useEffect(() => {
    if (!user.token) {
      navigate("/login");
      return;
    }
    dispatch(meApi(user.token));
    dispatch(listOrgsApi(user.token));
    // Depend on the token only: meApi syncs the profile (role, name, picture...)
    // into this same user object, so depending on the whole object would
    // re-trigger this effect on every successful sync, calling meApi again
    // in an infinite loop.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user.token, dispatch, navigate]);

  const handleLogout = () => {
    dispatch(logoutUser());
    navigate("/login");
  };

  if (user.role === "disabled") {
    return <DisabledNotice />;
  }

  return (
    <div className={styles.main}>
      <div className={styles.navbar}>
        <div className={styles.navbarleftmain}>
          <img src={logo} alt="cwclock logo" className={styles.logo} title="CWClock" />

          {organizations.length > 0 ? (
            <Dropdown
              title="Switch organization"
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
            <Tooltip label="Account menu">
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
              {(close) => (
                <>
                  <DropdownText>{memberLabel(user)}</DropdownText>
                  {(user.name || user.surname) && (
                    <DropdownText className={styles.profileEmail}>{user.email}</DropdownText>
                  )}
                  <DropdownItem
                    onClick={() => {
                      setShowEditProfile(true);
                      close();
                    }}
                    title="Edit your profile and avatar"
                  >
                    Edit profile
                  </DropdownItem>
                  <DropdownDivider />
                  <Tooltip label="Logout" className={styles.logoutTooltip}>
                    <DropdownItem onClick={handleLogout} className={styles.logoutItem}>
                      <FiLogOut style={{ fontSize: "16px" }} />
                    </DropdownItem>
                  </Tooltip>
                </>
              )}
            </Dropdown>
            </Tooltip>
          ) : (
            <button className={styles.loginBtn} onClick={() => navigate("/login")} title="Go to the login page">
              Login
            </button>
          )}
        </div>
      </div>

      <EditProfileModal show={showEditProfile} onClose={() => setShowEditProfile(false)} user={user} />

      <div className={styles.Slideflex}>
        <div className={`${styles.sidebarCol} ${expanded ? styles.sidebarColExpanded : ""}`}>
          <SidebarNav expanded={expanded} isSuperuser={isSuperuser} />
          <Tooltip label={expanded ? "Collapse sidebar" : "Expand sidebar"} position="right" className={styles.toggleTooltip}>
            <button className={styles.sidebarToggle} onClick={handleclick}>
              {expanded ? <FaChevronLeft /> : <FaChevronRight />}
            </button>
          </Tooltip>
        </div>
        <div className={styles.pages}>
          <Routes>
            <Route path="/timetracker" element={<TimeTracker />}></Route>
            <Route path="/organizations" element={<Organizationsdiv />}></Route>
            <Route path="/clients" element={<Clientdiv />}></Route>
            <Route path="/projects" element={<Projectsdiv />}></Route>
            {isSuperuser && <Route path="/admin" element={<Admindiv />}></Route>}
            {isSuperuser && (
              <Route path="/admin/organizations" element={<AdminOrganizationsdiv />}></Route>
            )}
          </Routes>
        </div>
      </div>
    </div>
  );
};

export default Slidebar;
