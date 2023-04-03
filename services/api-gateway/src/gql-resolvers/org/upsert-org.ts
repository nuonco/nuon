import { UpsertOrgRequest } from "@buf/nuon_apis.grpc_node/org/v1/messages_pb";
import { GraphQLError } from "graphql";
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
