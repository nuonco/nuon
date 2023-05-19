import { builds } from "./builds";
import { cancelBuild } from "./cancel-build";
import { startBuild } from "./start-build";

export const buildResolvers = {
  Mutation: {
    cancelBuild,
    startBuild,
  },
  Query: {
    builds,
  },
};
