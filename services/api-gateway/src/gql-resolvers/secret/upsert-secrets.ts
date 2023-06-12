import {
  Secret,
  UpsertSecretsRequest,
} from "@buf/nuon_orgs-api.grpc_node/instances/v1/secrets_pb";
import { GraphQLError } from "graphql";
import type {
  Mutation,
  MutationUpsertSecretsArgs,
  SecretsInput,
  TResolverFn,
} from "../../types";

export function parseSecretInput(input: SecretsInput[]) {
  return input.flatMap(({ appId, componentId, installId, orgId, secrets }) => {
    return secrets.map(({ id, key, value }) => {
      const secret = new Secret()
        .setAppId(appId)
        .setOrgId(orgId)
        .setComponentId(componentId)
        .setInstallId(installId)
        .setId(id)
        .setKey(key)
        .setValue(value);
      return secret;
    });
  });
}

export const upsertSecrets: TResolverFn<
  MutationUpsertSecretsArgs,
  Mutation["upsertSecrets"]
> = (_, { input }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.instanceStatus) {
      const request = new UpsertSecretsRequest().setSecretsList(
        parseSecretInput(input)
      );

      clients.instanceStatus.upsertSecrets(request, (err) => {
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
