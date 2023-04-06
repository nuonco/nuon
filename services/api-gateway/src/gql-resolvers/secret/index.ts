import { secrets } from "./secrets";
import { upsertSecrets } from "./upsert-secrets";

export const secretResolvers = {
  Mutation: {
    upsertSecrets,
  },
  Query: {
    secrets,
  },
};
