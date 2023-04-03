import { GetAppRequest } from "@buf/nuon_apis.grpc_node/app/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TApp, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const app: TResolverFn<{ id: string }, TApp> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new GetAppRequest().setId(id);

      clients.app.getApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(getNodeFields<TApp>(res.toObject().app));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
