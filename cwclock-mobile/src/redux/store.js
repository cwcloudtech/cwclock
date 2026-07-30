import { createStore, combineReducers, applyMiddleware } from "redux";
import thunk from "redux-thunk";
import { sessionReducer } from "./session/session.reducer";
import { organizationsReducer } from "./organizations/organizations.reducer";
import { clientsReducer } from "./clients/clients.reducer";
import { projectsReducer } from "./projects/projects.reducer";
import { timeEntriesReducer } from "./timeEntries/timeEntries.reducer";
import { invoicesReducer } from "./invoices/invoices.reducer";

const rootReducer = combineReducers({
  session: sessionReducer,
  organizations: organizationsReducer,
  clients: clientsReducer,
  projects: projectsReducer,
  timeEntries: timeEntriesReducer,
  invoices: invoicesReducer,
});

const store = createStore(rootReducer, applyMiddleware(thunk));

export default store;
