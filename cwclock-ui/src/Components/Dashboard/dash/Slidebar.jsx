import React, { useState, useEffect } from "react";
import styles from "./STYLE/Slidebar.module.css";
import { FaChevronLeft, FaChevronRight, FaUserCheck } from "react-icons/fa";
import { Route, Routes, useNavigate, Link } from "react-router-dom";
import Dropdown from "react-bootstrap/Dropdown";
import logo from "../../../assets/images/cwclock-logo.svg";
import TimeTracker from "../pages/TimeTracker";
import Slideclose from "./Slideclose";
import Slideopen from "./Slideopen";

import Clientdiv from "../pages/Client";
import Organizationsdiv from "../pages/Organizations";
import Projectsdiv from "../pages/Projects";
import { useSelector, useDispatch } from "react-redux";
import { meApi, logoutUser } from "../../../Redux/Auth/Auth.actions";
import { updatePictureApi } from "../../../Redux/Users/User.actions";
import { listOrgsApi, selectOrg } from "../../../Redux/Organizations/Org.actions";
import fileToBase64 from "../../common/fileToBase64";

const Slidebar = () => {
  const [state, setstate] = useState(false);
  const handleclick = () => {
    setstate(!state);
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
          <div>
            <img src={logo} alt="cwclock logo"></img>
          </div>

          {organizations.length > 0 ? (
            <Dropdown>
              <Dropdown.Toggle variant="light" className={styles.orgDropdownToggle}>
                {currentOrg?.picture && (
                  <img src={currentOrg.picture} alt="" className={styles.orgAvatar} />
                )}
                {currentOrg?.name}
              </Dropdown.Toggle>
              <Dropdown.Menu>
                {organizations.map((org) => (
                  <Dropdown.Item
                    key={org.id}
                    onClick={() => dispatch(selectOrg(org.id))}
                    active={org.id === currentOrgId}
                  >
                    {org.picture && <img src={org.picture} alt="" className={styles.orgAvatar} />}
                    {org.name}
                  </Dropdown.Item>
                ))}
              </Dropdown.Menu>
            </Dropdown>
          ) : (
            <Link className={styles.createOrgLink} to="/dashboard/organizations">
              Create organization
            </Link>
          )}
        </div>

        {/* right div */}
        <div className={styles.navbarrightmain}>
          {user.token ? (
            <Dropdown align="end">
              <Dropdown.Toggle variant="light" className={styles.profileDropdownToggle}>
                {user.picture ? (
                  <img src={user.picture} alt="" className={styles.profileAvatar} />
                ) : (
                  <FaUserCheck fontSize="22px" />
                )}
              </Dropdown.Toggle>
              <Dropdown.Menu>
                <Dropdown.ItemText className={styles.profileEmail}>{user.email}</Dropdown.ItemText>
                <Dropdown.ItemText>
                  <label className={styles.uploadLabel}>
                    Change picture
                    <input type="file" accept="image/*" hidden onChange={handlePictureChange} />
                  </label>
                </Dropdown.ItemText>
                <Dropdown.Divider />
                <Dropdown.Item onClick={handleLogout}>Logout</Dropdown.Item>
              </Dropdown.Menu>
            </Dropdown>
          ) : (
            <button className={styles.loginBtn} onClick={() => navigate("/login")}>
              Login
            </button>
          )}
        </div>
      </div>

      <div className={styles.Slideflex}>
        <div className={styles.sidebarCol}>
          <div
            className={styles.sidebarToggle}
            onClick={handleclick}
            title={state ? "Collapse sidebar" : "Expand sidebar"}
          >
            {state ? <FaChevronLeft /> : <FaChevronRight />}
          </div>
          {state ? (
            <div className={styles.slidingfuncbox}>
              <div>
                <Slideopen />
              </div>
              <div>
                <Slideclose />
              </div>
            </div>
          ) : (
            <div>
              <Slideopen />
            </div>
          )}
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
