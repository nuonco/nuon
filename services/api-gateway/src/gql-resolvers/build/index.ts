import { build } from "./build";
import { buildStatus } from "./build-status";
import { builds } from "./builds";
import { cancelBuild } from "./cancel-build";
import { startBuild } from "./start-build";

export const buildResolvers = {
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
