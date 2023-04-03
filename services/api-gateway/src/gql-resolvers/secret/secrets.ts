import { GetSecretsRequest } from "@buf/nuon_orgs-api.grpc_node/instances/v1/secrets_pb";
import { GraphQLError } from "graphql";
import { TResolverFn, TSecret } from "../../types";

interface ISecretsResolver {
  appId: string;
  componentId: string;
  installId: string;
  orgId: string;
}

export const secrets: TResolverFn<ISecretsResolver, TSecret[]> = (
  _,
  { appId, componentId, installId, orgId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.instance) {
      const request = new GetSecretsRequest()
        .setAppId(appId)
        .setComponentId(componentId)
        .setInstallId(installId)
        .setOrgId(orgId);

      clients.instance.getSecrets(request, (err, res) => {
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
