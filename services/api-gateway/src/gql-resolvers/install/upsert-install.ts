import {
  AwsRegion,
  AwsSettings,
  UpsertInstallRequest,
} from "@buf/nuon_apis.grpc_node/install/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TInstall, TResolverFn } from "../../types";
import { formatInstall } from "./utils";

type TInstallInput = {
  appId?: string;
  awsSettings: {
    region: "US_EAST_1" | "US_EAST_2" | "US_WEST_1" | "US_WEST_2";
    role: string;
  };
  id?: string;
  name?: string;
};

export const upsertInstall: TResolverFn<{ input: TInstallInput }, TInstall> = (
  _,
  { input },
  { clients, user }
) =>
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
