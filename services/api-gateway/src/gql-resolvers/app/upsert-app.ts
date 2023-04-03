import { UpsertAppRequest } from "@buf/nuon_apis.grpc_node/app/v1/messages_pb";
import { GraphQLError } from "graphql";
import { TApp, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

type TAppInput = {
  githubInstallId?: string;
  id?: string;
  name?: string;
  orgId?: string;
};

export const upsertApp: TResolverFn<{ input: TAppInput }, TApp> = (
  _,
  { input },
  { clients, user }
) =>
  new Promise((resolve, reject) => {
    if (clients.app) {
      const request = new UpsertAppRequest()
        .setId(input.id)
        .setOrgId(input.orgId)
        .setName(input.name)
        .setGithubInstallId(input.githubInstallId)
        .setCreatedById(user?.id);

      clients.app.upsertApp(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(getNodeFields(res.toObject().app));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
