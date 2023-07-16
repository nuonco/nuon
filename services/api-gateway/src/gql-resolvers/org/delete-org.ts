import { GraphQLError } from "graphql";
import { DeleteOrgRequest } from "../../build/api/org/v1/messages_pb";
import type { Mutation, MutationDeleteOrgArgs, TResolverFn } from "../../types";

export const deleteOrg: TResolverFn<
  MutationDeleteOrgArgs,
  Mutation["deleteOrg"]
> = (_, { id }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.org) {
      const request = new DeleteOrgRequest().setId(id);

      clients.org.deleteOrg(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(res.toObject().deleted);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
