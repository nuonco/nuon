import { GraphQLError } from "graphql";
import { GetAppRequest } from "../../build/api/app/v1/messages_pb";
import type { App, Query, QueryAppArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const app: TResolverFn<QueryAppArgs, Query["app"]> = (
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
          resolve(getNodeFields<App>(res.toObject().app));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
