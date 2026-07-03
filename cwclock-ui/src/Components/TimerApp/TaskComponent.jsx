import React, { useEffect, useState } from "react";
import styles from "./Styles/TaskComp.module.css";
import { MdDeleteForever } from "react-icons/md";
import { useDispatch, useSelector } from "react-redux";
import { deleteTasksApi, updateTasksApi } from "../../Redux/Tasks/Task.actions";
import ConfirmModal from "../common/ConfirmModal";

const fieldsFromItem = (item) => ({
  text: item.text,
  day: item.day,
  start: item.start || "",
  end: item.end || "",
  allDay: item.allDay,
  status: item.status,
  userId: item.userId,
});

const TaskComponent = ({ item }) => {
  const [isEditing, setIsEditing] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [form, setForm] = useState(fieldsFromItem(item));
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
    }
  }, [item, isEditing]);

  const handleDelete = () => {
    setShowDeleteConfirm(false);
    dispatch(deleteTasksApi(item.id, currentOrgId, user.token));
  };

  const handleSave = () => {
    const update = {
      ...item,
      text: form.text,
      day: form.day,
      status: form.status,
      allDay: form.allDay,
      start: form.allDay ? null : form.start,
      end: form.allDay ? null : form.end,
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
                value={form.text}
                onChange={(e) => setForm({ ...form, text: e.target.value })}
              />
            </div>
            <div className={styles.Projects}>
              <h6>{project ? project.name : ""}</h6>
            </div>
            <div className={styles.Time}>
              <input
                type="date"
                value={form.day}
                onChange={(e) => setForm({ ...form, day: e.target.value })}
              />
              <label>
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
                    value={form.start}
                    onChange={(e) => setForm({ ...form, start: e.target.value })}
                  />
                  <input
                    type="time"
                    step="1"
                    value={form.end}
                    onChange={(e) => setForm({ ...form, end: e.target.value })}
                  />
                </>
              )}
              <label>
                <input
                  type="checkbox"
                  checked={form.status}
                  onChange={(e) => setForm({ ...form, status: e.target.checked })}
                />{" "}
                Done
              </label>
              {isAdminOrOwner && (
                <select
                  value={form.userId}
                  onChange={(e) => setForm({ ...form, userId: e.target.value })}
                >
                  {members.map((m) => (
                    <option key={m.userId} value={m.userId}>
                      {m.email}
                    </option>
                  ))}
                </select>
              )}
            </div>
            <button type="button" className={styles.Tags3} onClick={handleSave}>
              Save
            </button>
            <button type="button" className={styles.Tags} onClick={() => setIsEditing(false)}>
              Cancel
            </button>
          </>
        ) : (
          <>
            <div
              className={styles.Up}
              onClick={canEdit ? () => setIsEditing(true) : undefined}
              style={canEdit ? undefined : { cursor: "default" }}
            >
              <h6 style={{ color: item.status ? "green" : "var(--cw-danger)" }}>{item.text}</h6>
            </div>
            <div className={styles.Projects}>
              <h6>{project ? project.name : ""}</h6>
            </div>
            <div className={styles.Time}>
              <span>{item.day}</span>
              <span>{timeLabel}</span>
              {assignee && <span>{assignee.email}</span>}
            </div>
          </>
        )}
        {!isEditing && canEdit && (
          <button
            type="button"
            className={styles.Tags2}
            onClick={() => setShowDeleteConfirm(true)}
          >
            <MdDeleteForever style={{ fontSize: "25px" }} />
          </button>
        )}
      </div>

      <ConfirmModal
        show={showDeleteConfirm}
        title="Delete time record"
        body={`Delete "${item.text}"? This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setShowDeleteConfirm(false)}
      />
    </div>
  );
};

export default TaskComponent;
