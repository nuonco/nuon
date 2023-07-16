import { GraphQLError } from "graphql";
import { GetComponentRequest } from "../../build/api/component/v1/messages_pb";
import type { Query, QueryComponentArgs, TResolverFn } from "../../types";
import { formatComponent } from "./utils";

export const component: TResolverFn<QueryComponentArgs, Query["component"]> = (
  _,
  { id },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.component) {
      const request = new GetComponentRequest().setId(id);

      clients.component.getComponent(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(formatComponent(res.toObject().component));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
