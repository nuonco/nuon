import {
  DeleteSecretsRequest,
  SecretRef,
} from "@buf/nuon_orgs-api.grpc_node/instances/v1/secrets_pb";
import { GraphQLError } from "graphql";
import type {
  Mutation,
  MutationDeleteSecretsArgs,
  SecretsIdsInput,
  TResolverFn,
} from "../../types";

export function parseSecretInput(input: SecretsIdsInput[]) {
  return input.flatMap(({ appId, componentId, installId, orgId, secretId }) => {
    const secretRef = new SecretRef()
      .setAppId(appId)
      .setOrgId(orgId)
      .setComponentId(componentId)
      .setInstallId(installId)
      .setSecretId(secretId);
    return secretRef;
  });
}

export const deleteSecrets: TResolverFn<
  MutationDeleteSecretsArgs,
  Mutation["deleteSecrets"]
> = (_, { input }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.instanceStatus) {
      const request = new DeleteSecretsRequest().setSecretRefsList(
        parseSecretInput(input)
      );

      clients.instanceStatus.deleteSecrets(request, (err) => {
        if (err) {
          reject(new GraphQLError(err?.message));
        } else {
          resolve(true);
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
