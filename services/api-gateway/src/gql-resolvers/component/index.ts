import { app } from "../app/app";
import { builds } from "../build/builds";
import { component } from "./component";
import { components } from "./components";
import { deleteComponent } from "./delete-component";
import { upsertComponent } from "./upsert-component";

export const componentResolvers = {
  Component: {
    app: (parent, _, ctx) => app(parent, { id: parent.appId }, ctx),
    builds: (parent, _, ctx) => builds(parent, { componentId: parent.id }, ctx),
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
