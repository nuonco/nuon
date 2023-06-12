import { GetSecretsRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/secrets_pb";
import { GraphQLError } from "graphql";
import type { Query, QuerySecretsArgs, TResolverFn } from "../../types";

export const secrets: TResolverFn<QuerySecretsArgs, Query["secrets"]> = (
  _,
  { appId, componentId, installId, orgId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.instanceStatus) {
      const request = new GetSecretsRequest()
        .setAppId(appId)
        .setComponentId(componentId)
        .setInstallId(installId)
        .setOrgId(orgId);

      clients.instanceStatus.getSecrets(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          const { secretsList } = res.toObject();
          resolve(secretsList || []);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
