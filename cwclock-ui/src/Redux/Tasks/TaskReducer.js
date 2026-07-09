import {
  DeleteTasksSUCCESS,
  GetTasksERROR,
  GetTasksLOADING,
  GetTasksLOADINGMORE,
  GetTasksSUCCESS,
  GetTasksAppendSUCCESS,
  PostTasksSUCCESS,
  start,
  UpdateTasksSUCCESS,
} from "./Task.types";

// Matches the API's own ORDER BY data->>'day' DESC, data->>'start' DESC, so
// a task inserted/edited client-side (rather than re-fetched) lands in the
// same slot the server would put it in - TasksApp's day-grouping assumes
// the list is already in this order and just walks it once.
const compareEntries = (a, b) => {
  if (a.day !== b.day) return a.day > b.day ? -1 : 1;
  const aStart = a.start || "";
  const bStart = b.start || "";
  if (aStart === bStart) return 0;
  return aStart > bStart ? -1 : 1;
};

const initialstate = {
  tasks: [],
  start: "",
  isLoading: false,
  isLoadingMore: false,
  isError: false,
  page: 1,
  hasMore: false,
};
export const TaskReducer = (state = initialstate, { type, payload }) => {
  switch (type) {
    case start: {
      return { ...state, start: payload };
    }
    case GetTasksLOADING: {
      return { ...state, isLoading: true };
    }
    case GetTasksLOADINGMORE: {
      return { ...state, isLoadingMore: true };
    }
    case GetTasksERROR: {
      return { ...state, isLoading: false, isLoadingMore: false, isError: true };
    }
    case GetTasksSUCCESS: {
      const items = Array.isArray(payload?.items) ? payload.items : [];
      return { ...state, tasks: items, isLoading: false, page: payload?.page ?? 1, hasMore: !!payload?.hasMore };
    }
    case GetTasksAppendSUCCESS: {
      const items = Array.isArray(payload?.items) ? payload.items : [];
      return {
        ...state,
        tasks: [...state.tasks, ...items],
        isLoadingMore: false,
        page: payload?.page ?? state.page,
        hasMore: !!payload?.hasMore,
      };
    }
    case PostTasksSUCCESS: {
      return { ...state, tasks: [...state.tasks, payload].sort(compareEntries), isLoading: false };
    }
    case DeleteTasksSUCCESS: {
      let filtered = state.tasks.filter((item) => {
        return item.id !== payload;
      });
      return { ...state, tasks: [...filtered], isLoading: false };
    }
    case UpdateTasksSUCCESS: {
      const updated = state.tasks.map((task) => (task.id === payload.id ? { ...payload } : task));
      return { ...state, tasks: updated.sort(compareEntries), isLoading: false };
    }
    default: {
      return state;
    }
  }
};
