import { DeleteOrgRequest } from "@buf/nuon_apis.grpc_node/org/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TResolverFn } from "../../types";

export const deleteOrg: TResolverFn<{ id: string }, boolean> = (
  _,
  { id },
  { clients }
) =>
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
