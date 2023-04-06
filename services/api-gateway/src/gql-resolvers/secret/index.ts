import { deleteSecrets } from "./delete-secrets";
import { secrets } from "./secrets";
import { upsertSecrets } from "./upsert-secrets";

export const secretResolvers = {
  Mutation: {
    deleteSecrets,
    upsertSecrets,
  },
  Query: {
    secrets,
  },
};
