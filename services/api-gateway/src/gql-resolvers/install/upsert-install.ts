import { GraphQLError } from "graphql";
import {
  AwsRegion,
  AwsSettings,
  UpsertInstallRequest,
} from "../../build/api/install/v1/messages_pb";
import type {
  Mutation,
  MutationUpsertInstallArgs,
  TResolverFn,
} from "../../types";
import { formatInstall } from "./utils";

export const upsertInstall: TResolverFn<
  MutationUpsertInstallArgs,
  Mutation["upsertInstall"]
> = (_, { input }, { clients, user }) =>
  new Promise((resolve, reject) => {
    if (clients.install) {
      const awsSettingsMessage = new AwsSettings()
        .setRegion(AwsRegion[`AWS_REGION_${input?.awsSettings?.region}`])
        .setRole(input?.awsSettings?.role);

      const request = new UpsertInstallRequest()
        .setId(input.id)
        .setAppId(input.appId)
        .setName(input.name)
        .setAwsSettings(awsSettingsMessage)
        .setCreatedById(user?.id);

      clients.install.upsertInstall(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(formatInstall(res.toObject().install));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
