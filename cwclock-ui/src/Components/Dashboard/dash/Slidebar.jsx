import React, { useState, useEffect } from "react";
import styles from "./STYLE/Slidebar.module.css";
import { FaChevronLeft, FaChevronRight, FaUserCheck } from "react-icons/fa";
import { FiLogOut, FiSun, FiMoon, FiGlobe } from "react-icons/fi";
import { Route, Routes, useNavigate, Link } from "react-router-dom";
import Dropdown, { DropdownItem, DropdownText, DropdownDivider } from "../../common/Dropdown";
import EditProfileModal from "../../common/EditProfileModal";
import DisabledNotice from "../../common/DisabledNotice";
import memberLabel from "../../common/memberLabel";
import Tooltip from "../../common/Tooltip";
import { useTheme } from "../../common/ThemeContext";
import { useI18n, LANGUAGES } from "../../../i18n/I18nContext";
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
  const { theme, toggleTheme } = useTheme();
  const { locale, setLocale, t } = useI18n();
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
          <img src={logo} alt="cwclock logo" className={styles.logo} title={t("nav.cwclock")} />

          {organizations.length > 0 ? (
            <Dropdown
              title={t("nav.switchOrganization")}
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
              {t("nav.createOrganization")}
            </Link>
          )}
        </div>

        <div className={styles.navbarrightmain}>
          {user.token ? (
            <Tooltip label={t("nav.accountMenu")}>
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
                    title={t("nav.editProfileTooltip")}
                  >
                    {t("nav.editProfile")}
                  </DropdownItem>
                  <Tooltip
                    label={theme === "dark" ? t("nav.switchToLightMode") : t("nav.switchToDarkMode")}
                    className={styles.menuItemTooltip}
                  >
                    <DropdownItem onClick={toggleTheme}>
                      {theme === "dark" ? (
                        <FiSun style={{ fontSize: "16px" }} />
                      ) : (
                        <FiMoon style={{ fontSize: "16px" }} />
                      )}
                      {theme === "dark" ? t("nav.lightMode") : t("nav.darkMode")}
                    </DropdownItem>
                  </Tooltip>
                  <DropdownText className={styles.languageLabel}>
                    <FiGlobe style={{ fontSize: "14px" }} /> {t("common.language")}
                  </DropdownText>
                  {LANGUAGES.map((lang) => (
                    <DropdownItem
                      key={lang.code}
                      active={lang.code === locale}
                      onClick={() => setLocale(lang.code)}
                      title={lang.label}
                    >
                      <span aria-hidden="true">{lang.flag}</span> {lang.label}
                    </DropdownItem>
                  ))}
                  <DropdownDivider />
                  <Tooltip label={t("nav.logout")} className={styles.menuItemTooltip}>
                    <DropdownItem onClick={handleLogout} className={styles.logoutItem}>
                      <FiLogOut style={{ fontSize: "16px" }} />
                    </DropdownItem>
                  </Tooltip>
                </>
              )}
            </Dropdown>
            </Tooltip>
          ) : (
            <button className={styles.loginBtn} onClick={() => navigate("/login")} title={t("nav.goToLogin")}>
              {t("nav.login")}
            </button>
          )}
        </div>
      </div>

      <EditProfileModal show={showEditProfile} onClose={() => setShowEditProfile(false)} user={user} />

      <div className={styles.Slideflex}>
        <div className={`${styles.sidebarCol} ${expanded ? styles.sidebarColExpanded : ""}`}>
          <SidebarNav expanded={expanded} isSuperuser={isSuperuser} />
          <Tooltip label={expanded ? t("nav.collapseSidebar") : t("nav.expandSidebar")} position="right" className={styles.toggleTooltip}>
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
