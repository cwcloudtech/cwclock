import React, { useState, useEffect } from "react";
import axios from "axios";
import styles from "./STYLE/Slidebar.module.css";
import { FaChevronLeft, FaChevronRight, FaUserCheck, FaGitAlt, FaBook, FaEnvelope } from "react-icons/fa";
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
import Reportsdiv from "../pages/Reports";
import Invoicesdiv from "../pages/Invoices";
import Admindiv from "../pages/Admin";
import AdminOrganizationsdiv from "../pages/AdminOrganizations";
import ApiKeysdiv from "../pages/ApiKeys";
import { useSelector, useDispatch } from "react-redux";
import { meApi, logoutUser } from "../../../Redux/Auth/Auth.actions";
import { listOrgsApi, selectOrg, listMembersApi } from "../../../Redux/Organizations/Org.actions";
import { isAdminOrOwner as computeIsAdminOrOwner } from "../../common/permissions";

const Slidebar = () => {
  const [expanded, setExpanded] = useState(false);
  const [showEditProfile, setShowEditProfile] = useState(false);
  const [appVersion, setAppVersion] = useState(null);

  useEffect(() => {
    axios
      .get(`${process.env.REACT_APP_APIURL}/v1/manifest`)
      .then(({ data }) => setAppVersion(data.version))
      .catch(() => {});
  }, []);
  const handleclick = () => {
    setExpanded(!expanded);
  };

  const navigate = useNavigate();
  const dispatch = useDispatch();
  const { theme, toggleTheme } = useTheme();
  const { locale, setLocale, t } = useI18n();
  const { user } = useSelector((state) => state.auth);
  const { organizations, currentOrgId, members } = useSelector((state) => state.organizations);
  const currentOrg = organizations.find((o) => o.id === currentOrgId);
  const isSuperuser = user.role === "superuser";
  const showInvoices = computeIsAdminOrOwner(user, members);

  useEffect(() => {
    if (currentOrgId) {
      dispatch(listMembersApi(currentOrgId, user.token));
    }
  }, [dispatch, currentOrgId, user.token]);

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

  if (user.role === "disabled" || user.role === "ban") {
    return <DisabledNotice user={user} />;
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
                    <img
                      src={currentOrg.picture}
                      alt=""
                      className={styles.orgAvatar}
                      style={{ objectPosition: `${currentOrg.pictureX ?? 50}% ${currentOrg.pictureY ?? 50}%` }}
                    />
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
                    {org.picture && (
                      <img
                        src={org.picture}
                        alt=""
                        className={styles.orgAvatar}
                        style={{ objectPosition: `${org.pictureX ?? 50}% ${org.pictureY ?? 50}%` }}
                      />
                    )}
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
            <>
            <Tooltip label={t("nav.documentation")} position="bottom">
              <a
                href={process.env[`REACT_APP_${locale.toUpperCase()}_DOCURL`] || process.env.REACT_APP_EN_DOCURL}
                target="_blank"
                rel="noopener noreferrer"
                className={styles.gitRepoLink}
              >
                <FaBook fontSize="18px" />
              </a>
            </Tooltip>
            <Tooltip label={t("nav.gitRepository")} position="bottom">
              <a
                href={process.env.REACT_APP_REPOURL}
                target="_blank"
                rel="noopener noreferrer"
                className={styles.gitRepoLink}
              >
                <FaGitAlt fontSize="18px" />
              </a>
            </Tooltip>
            <Tooltip label={t("nav.contactForm")} position="bottom">
              <Link to="/contact" className={styles.gitRepoLink}>
                <FaEnvelope fontSize="18px" />
              </Link>
            </Tooltip>
            <Dropdown
              align="end"
              triggerClassName={styles.profileTrigger}
              trigger={
                user.picture ? (
                  <img
                    src={user.picture}
                    alt=""
                    className={styles.profileAvatar}
                    style={{ objectPosition: `${user.pictureX ?? 50}% ${user.pictureY ?? 50}%` }}
                  />
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
                  {appVersion && (
                    <DropdownText className={styles.versionLabelWrap}>
                      <span className={styles.versionBadge}>v{appVersion}</span>
                    </DropdownText>
                  )}
                  <Tooltip label={t("nav.logout")} className={styles.menuItemTooltip}>
                    <DropdownItem onClick={handleLogout} className={styles.logoutItem}>
                      <FiLogOut style={{ fontSize: "16px" }} />
                    </DropdownItem>
                  </Tooltip>
                </>
              )}
            </Dropdown>
            </>
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
          <Tooltip label={expanded ? t("nav.collapseSidebar") : t("nav.expandSidebar")} position="right" className={styles.toggleTooltip}>
            <button className={styles.sidebarToggle} onClick={handleclick}>
              {expanded ? <FaChevronLeft /> : <FaChevronRight />}
            </button>
          </Tooltip>
          <SidebarNav expanded={expanded} isSuperuser={isSuperuser} showInvoices={showInvoices} />
        </div>
        <div className={styles.pages}>
          <Routes>
            <Route path="/timetracker" element={<TimeTracker />}></Route>
            <Route path="/reports" element={<Reportsdiv />}></Route>
            <Route path="/organizations" element={<Organizationsdiv />}></Route>
            <Route path="/clients" element={<Clientdiv />}></Route>
            <Route path="/projects" element={<Projectsdiv />}></Route>
            <Route path="/api-keys" element={<ApiKeysdiv />}></Route>
            {showInvoices && <Route path="/invoices" element={<Invoicesdiv />}></Route>}
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
