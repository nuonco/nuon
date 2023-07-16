import { GraphQLError } from "graphql";
import {
  GetDeployRequest,
  GetDeploysByInstanceRequest,
} from "../../build/api/deploy/v1/messages_pb";
import type { Deploy, Query, QueryDeployArgs, TResolverFn } from "../../types";
import { getNodeFields } from "../../utils";

export const deploy: TResolverFn<QueryDeployArgs, Query["deploy"]> = (
  _,
  { id, instanceId },
  { clients }
) =>
  new Promise((resolve, reject) => {
    if (clients.deploy) {
      if (!id && !instanceId) {
        reject(
          new GraphQLError(
            "Invalid query parameters, please provide either a deploy ID or instance ID"
          )
        );
      }

      if (id) {
        const request = new GetDeployRequest().setId(id);
        clients.deploy.getDeploy(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            resolve(getNodeFields<Deploy>(res.toObject().deploy));
          }
        });
      }

      if (instanceId) {
        const request = new GetDeploysByInstanceRequest().setInstanceId(
          instanceId
        );

        clients.deploy.getDeploysByInstance(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            const { deploysList } = res.toObject();
            const recentDeploy = deploysList.reverse()[0];
            resolve(recentDeploy ? getNodeFields(recentDeploy) : null);
          }
        });
      }
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
