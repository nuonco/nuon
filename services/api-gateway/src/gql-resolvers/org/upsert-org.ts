import { GraphQLError } from "graphql";
import { UpsertOrgRequest } from "../../build/api/org/v1/messages_pb";
import type {
  Mutation,
  MutationUpsertOrgArgs,
  Org,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

export const upsertOrg: TResolverFn<
  MutationUpsertOrgArgs,
  Mutation["upsertOrg"]
> = (_, { input }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.org) {
      const request = new UpsertOrgRequest()
        .setId(input.id)
        .setOwnerId(input.ownerId)
        .setGithubInstallId(input.githubInstallId)
        .setName(input.name);

      clients.org.upsertOrg(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(getNodeFields<Org>(res.toObject().org));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
