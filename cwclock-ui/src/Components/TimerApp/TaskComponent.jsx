import React, { useEffect, useState } from "react";
import styles from "./Styles/TaskComp.module.css";
import { MdDeleteForever } from "react-icons/md";
import { FaRegEdit } from "react-icons/fa";
import { useDispatch, useSelector } from "react-redux";
import { deleteTasksApi, updateTasksApi } from "../../Redux/Tasks/Task.actions";
import ConfirmModal from "../common/ConfirmModal";
import memberLabel from "../common/memberLabel";
import Tooltip from "../common/Tooltip";

const fieldsFromItem = (item) => ({
  text: item.text,
  day: item.day,
  start: item.start || "",
  end: item.end || "",
  allDay: item.allDay,
  projectId: item.projectId,
  userId: item.userId,
});

const TaskComponent = ({ item }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [form, setForm] = useState(fieldsFromItem(item));
  const [reassignText, setReassignText] = useState("");
  const { user } = useSelector((state) => state.auth);
  const { currentOrgId, members } = useSelector((state) => state.organizations);
  const { projects } = useSelector((state) => state.projects);
  const dispatch = useDispatch();

  const project = projects.find((p) => p.id === item.projectId);
  const myRole = members.find((m) => m.userId === user.id)?.role;
  const isAdminOrOwner = myRole === "admin" || myRole === "owner";
  const canEdit = myRole && myRole !== "reader";
  const assignee = members.find((m) => m.userId === item.userId);

  useEffect(() => {
    if (!isEditing) {
      setForm(fieldsFromItem(item));
    } else {
      const current = members.find((m) => m.userId === item.userId);
      setReassignText(current ? memberLabel(current) : "");
    }
  }, [item, isEditing, members]);

  const handleReassignInput = (text) => {
    setReassignText(text);
    const match = members.find((m) => memberLabel(m) === text || m.email === text);
    if (match) {
      setForm((f) => ({ ...f, userId: match.userId }));
    }
  };

  const handleDelete = () => {
    setShowDeleteConfirm(false);
    dispatch(deleteTasksApi(item.id, currentOrgId, user.token));
  };

  const handleSave = () => {
    const selectedProject = projects.find((p) => p.id === form.projectId);
    const update = {
      ...item,
      text: form.text,
      day: form.day,
      allDay: form.allDay,
      start: form.allDay ? null : form.start,
      end: form.allDay ? null : form.end,
      projectId: form.projectId,
      clientId: selectedProject ? selectedProject.clientId : item.clientId,
      userId: form.userId,
    };
    dispatch(updateTasksApi(update, currentOrgId, user.token));
    setIsEditing(false);
  };

  const timeLabel = item.allDay ? "All day" : `${item.start || "?"} - ${item.end || "?"}`;

  return (
    <div className={styles.Task}>
      <div className={styles.Desc}>
        {isEditing ? (
          <>
            <div className={styles.Up}>
              <input
                type="text"
                autoFocus
                title="Task description"
                value={form.text}
                onChange={(e) => setForm({ ...form, text: e.target.value })}
              />
            </div>
            <div className={styles.Projects}>
              <select
                title="Project"
                value={form.projectId}
                onChange={(e) => setForm({ ...form, projectId: e.target.value })}
              >
                {projects.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>
            <div className={styles.Time}>
              <input
                type="date"
                title="Day"
                value={form.day}
                onChange={(e) => setForm({ ...form, day: e.target.value })}
              />
              <label title="Mark as an all-day entry">
                <input
                  type="checkbox"
                  checked={form.allDay}
                  onChange={(e) => setForm({ ...form, allDay: e.target.checked })}
                />{" "}
                All day
              </label>
              {!form.allDay && (
                <>
                  <input
                    type="time"
                    step="1"
                    title="Start time"
                    value={form.start}
                    onChange={(e) => setForm({ ...form, start: e.target.value })}
                  />
                  <input
                    type="time"
                    step="1"
                    title="End time"
                    value={form.end}
                    onChange={(e) => setForm({ ...form, end: e.target.value })}
                  />
                </>
              )}
              {isAdminOrOwner && (
                <>
                  <input
                    list={`reassign-options-${item.id}`}
                    title="Reassign to a member"
                    placeholder="Search member..."
                    value={reassignText}
                    onChange={(e) => handleReassignInput(e.target.value)}
                  />
                  <datalist id={`reassign-options-${item.id}`}>
                    {members.map((m) => (
                      <option key={m.userId} value={memberLabel(m)} />
                    ))}
                  </datalist>
                </>
              )}
            </div>
            <button type="button" className={styles.Tags3} onClick={handleSave} title="Save changes">
              Save
            </button>
            <button
              type="button"
              className={styles.Tags}
              onClick={() => setIsEditing(false)}
              title="Discard changes"
            >
              Cancel
            </button>
          </>
        ) : (
          <>
            <div className={styles.Up}>
              <h6>{item.text}</h6>
            </div>
            <div className={styles.Projects}>
              <h6>{project ? project.name : ""}</h6>
            </div>
            <div className={styles.Time}>
              <span>{item.day}</span>
              <span>{timeLabel}</span>
              {assignee && <span>{memberLabel(assignee)}</span>}
            </div>
          </>
        )}
        {!isEditing && canEdit && (
          <div className={styles.RowActions}>
            <Tooltip label="Edit">
              <button type="button" className={styles.Tags} onClick={() => setIsEditing(true)}>
                <FaRegEdit style={{ fontSize: "18px" }} />
              </button>
            </Tooltip>
            <Tooltip label="Delete">
              <button
                type="button"
                className={styles.Tags2}
                onClick={() => setShowDeleteConfirm(true)}
              >
                <MdDeleteForever style={{ fontSize: "20px" }} />
              </button>
            </Tooltip>
          </div>
        )}
      </div>

      <ConfirmModal
        show={showDeleteConfirm}
        title="Delete time record"
        body={`Are you sure to delete "${item.text}"? This can't be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </div>
  );
};

export default TaskComponent;
