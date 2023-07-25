import { app } from "../app/app";
import { components } from "../component/components";
import { deployments } from "../deployment/deployments";
import { deleteInstall } from "./delete-install";
import { install } from "./install";
import { installStatus } from "./install-status";
import { installs } from "./installs";
import { upsertInstall } from "./upsert-install";

export const installResolvers = {
  Install: {
    app: (parent, _, ctx) => app(parent, { id: parent.appId }, ctx),
    components: (parent, { options }, ctx) =>
      components(parent, { appId: parent.appId, options }, ctx),
    deployments: (parent, { options }, ctx) =>
      deployments(parent, { installIds: [parent.id], options }, ctx),
  },
  Mutation: {
    deleteInstall,
    upsertInstall,
  },
  Query: {
    install,
    installs,
    installStatus,
  },
};
