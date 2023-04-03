import { DeleteComponentRequest } from "@buf/nuon_apis.grpc_node/component/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TResolverFn } from "../../types";

export const deleteComponent: TResolverFn<{ id: string }, boolean> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.component) {
      const request = new DeleteComponentRequest().setId(id);

      clients.component.deleteComponent(request, (err, res) => {
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
