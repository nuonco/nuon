import { components } from "../component/components";
import { deployments } from "../deployment/deployments";
import { installs } from "../install/installs";
import { app } from "./app";
import { apps } from "./apps";
import { deleteApp } from "./delete-app";
import { upsertApp } from "./upsert-app";

export const appResolvers = {
  App: {
    components: (parent, { options }, ctx) =>
      components(parent, { appId: parent.id, options }, ctx),
    deployments: (parent, { options }, ctx) =>
      deployments(parent, { appIds: [parent.id], options }, ctx),
    installs: (parent, { options }, ctx) =>
      installs(parent, { appId: parent.id, options }, ctx),
  },
  Mutation: {
    deleteApp,
    upsertApp,
  },
  Query: {
    app,
    apps,
  },
};
