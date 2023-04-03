import { DeleteAppRequest } from "@buf/nuon_apis.grpc_node/app/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TResolverFn } from "../../types";

export const deleteApp: TResolverFn<{ id: string }, boolean> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new DeleteAppRequest().setId(id);

      clients.app.deleteApp(request, (err, res) => {
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
