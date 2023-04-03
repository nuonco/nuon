import { GetComponentsByAppRequest } from "@buf/nuon_apis.grpc_node/component/v1/messages_pb";
import { GraphQLError } from "graphql";
import type { Query, QueryComponentsArgs, TResolverFn } from "../../types";
import { formatComponent } from "./utils";

export const components: TResolverFn<
  QueryComponentsArgs,
  Query["components"]
> = (_, { appId }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.component) {
      const request = new GetComponentsByAppRequest().setAppId(appId);

      clients.component.getComponentsByApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { componentsList } = res.toObject();

          resolve({
            edges:
              componentsList?.map((component) => ({
                cursor: component?.id,
                node: formatComponent(component),
              })) || [],
            pageInfo: {
              endCursor: null,
              hasNextPage: false,
              hasPreviousPage: false,
              startCursor: null,
            },
            totalCount: componentsList?.length || 0,
          });
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
