import { app } from "../app/app";
import { deployments } from "../deployment/deployments";
import { component } from "./component";
import { components } from "./components";
import { deleteComponent } from "./delete-component";
import { upsertComponent } from "./upsert-component";

export const componentResolvers = {
  Component: {
    app: (parent, _, ctx) => app(parent, { id: parent.appId }, ctx),
    deployments: (parent, { options }, ctx) =>
      deployments(parent, { componentIds: [parent.id], options }, ctx),
  },
  Mutation: {
    deleteComponent,
    upsertComponent,
  },
  Query: {
    component,
    components,
  },
};
