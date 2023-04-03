import { UpsertOrgRequest } from "@buf/nuon_apis.grpc_node/org/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TOrg, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

type TOrgInput = {
  id?: string;
  name?: string;
  ownerId?: string;
};

export const upsertOrg: TResolverFn<{ input: TOrgInput }, TOrg> = (
  _,
  { input },
  { clients }
) =>
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
          resolve(getNodeFields(res.toObject().org));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
