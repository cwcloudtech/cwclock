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
      return { ...state, tasks: [...state.tasks, payload], isLoading: false };
    }
    case DeleteTasksSUCCESS: {
      let filtered = state.tasks.filter((item) => {
        return item.id !== payload;
      });
      return { ...state, tasks: [...filtered], isLoading: false };
    }
    case UpdateTasksSUCCESS: {
      let updated = state.tasks.map((task) => {
        if (task.id === payload.id) {
          return { ...payload };
        } else {
          return task;
        }
      });
      return { ...state, tasks: [...updated], isLoading: false };
    }
    default: {
      return state;
    }
  }
};
