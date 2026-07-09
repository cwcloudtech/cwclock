import { legacy_createStore, combineReducers, applyMiddleware } from "redux";
import { TaskReducer } from "./Tasks/TaskReducer";
import { AuthReducer } from "./Auth/Auth.Reducer";
import { OrgReducer } from "./Organizations/Org.reducer";
import { ClientReducer } from "./Clients/Client.reducer";
import { ProjectReducer } from "./Projects/Project.reducer";
import { AdminReducer } from "./Admin/Admin.reducer";
import { ReportReducer } from "./Reports/Report.reducer";
import { ApiKeyReducer } from "./ApiKeys/ApiKey.reducer";
import { InvoiceReducer } from "./Invoices/Invoice.reducer";
import thunk from "redux-thunk";

const rootRuducer = combineReducers({
  tasks: TaskReducer,
  auth: AuthReducer,
  organizations: OrgReducer,
  clients: ClientReducer,
  projects: ProjectReducer,
  admin: AdminReducer,
  reports: ReportReducer,
  apiKeys: ApiKeyReducer,
  invoices: InvoiceReducer,
});

export const Store = legacy_createStore(rootRuducer, applyMiddleware(thunk));
