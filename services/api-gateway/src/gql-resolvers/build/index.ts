import { component } from "../component/component";
import { build } from "./build";
import { buildStatus } from "./build-status";
import { builds } from "./builds";
import { cancelBuild } from "./cancel-build";
import { startBuild } from "./start-build";

export const buildResolvers = {
  Build: {
    component: (parent, _, ctx) =>
      component(parent, { id: parent.componentId }, ctx),
  },
  Mutation: {
    cancelBuild,
    startBuild,
  },
  Query: {
    build,
    builds,
    buildStatus,
  },
};
