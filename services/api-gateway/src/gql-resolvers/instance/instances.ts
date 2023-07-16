import { GraphQLError } from "graphql";
import { GetInstancesByInstallRequest } from "../../build/api/instance/v1/messages_pb";
import type { Query, QueryInstancesArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const instances: TResolverFn<QueryInstancesArgs, Query["instances"]> = (
  _,
  { installId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.instance) {
      const request = new GetInstancesByInstallRequest().setInstallId(
        installId
      );

      clients.instance.getInstancesByInstall(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { instancesList } = res.toObject();
          resolve(instancesList.reverse().map(getNodeFields) || []);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
