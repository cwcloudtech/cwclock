import { legacy_createStore, combineReducers, applyMiddleware } from "redux";
import { TaskReducer } from "./Tasks/TaskReducer";
import { AuthReducer } from "./Auth/Auth.Reducer";
import { OrgReducer } from "./Organizations/Org.reducer";
import { ClientReducer } from "./Clients/Client.reducer";
import { ProjectReducer } from "./Projects/Project.reducer";
import thunk from "redux-thunk";

const rootRuducer = combineReducers({
  tasks: TaskReducer,
  auth: AuthReducer,
  organizations: OrgReducer,
  clients: ClientReducer,
  projects: ProjectReducer,
});

export const Store = legacy_createStore(rootRuducer, applyMiddleware(thunk));
