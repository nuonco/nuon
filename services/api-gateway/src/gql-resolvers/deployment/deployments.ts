import { GraphQLError } from "graphql";
import {
  GetDeploymentsByAppsRequest,
  GetDeploymentsByComponentsRequest,
  GetDeploymentsByInstallsRequest,
} from "../../build/api/deployment/v1/messages_pb";
import type {
  Deployment,
  Query,
  QueryDeploymentsArgs,
  TResolverFn,
} from "../../types";
import { getNodeFields } from "../../utils";

export const deployments: TResolverFn<
  QueryDeploymentsArgs,
  Query["deployments"]
> = (_, { appIds, componentIds, installIds }, { clients }) =>
  new Promise((resolve, reject) => {
    if (clients.deployment) {
      if (componentIds) {
        const request =
          new GetDeploymentsByComponentsRequest().setComponentIdsList(
            componentIds
          );

        clients.deployment.getDeploymentsByComponents(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            const { deploymentsList } = res.toObject();

            resolve({
              edges:
                deploymentsList?.map((deployment) => ({
                  cursor: deployment?.id,
                  node: getNodeFields<Deployment>(deployment),
                })) || [],
              pageInfo: {
                endCursor: null,
                hasNextPage: false,
                hasPreviousPage: false,
                startCursor: null,
              },
              totalCount: deploymentsList.length || 0,
            });
          }
        });
      } else if (installIds) {
        const request = new GetDeploymentsByInstallsRequest().setInstallIdsList(
          installIds
        );

        clients.deployment.getDeploymentsByInstalls(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            const { deploymentsList } = res.toObject();

            resolve({
              edges:
                deploymentsList?.map((deployment) => ({
                  cursor: deployment?.id,
                  node: getNodeFields<Deployment>(deployment),
                })) || [],
              pageInfo: {
                endCursor: null,
                hasNextPage: false,
                hasPreviousPage: false,
                startCursor: null,
              },
              totalCount: deploymentsList.length || 0,
            });
          }
        });
      } else if (appIds) {
        const request = new GetDeploymentsByAppsRequest().setAppIdsList(appIds);

        clients.deployment.getDeploymentsByApps(request, (err, res) => {
          if (err) {
            reject(new GraphQLError(err?.message));
          } else {
            const { deploymentsList } = res.toObject();

            resolve({
              edges:
                deploymentsList?.map((deployment) => ({
                  cursor: deployment?.id,
                  node: getNodeFields<Deployment>(deployment),
                })) || [],
              pageInfo: {
                endCursor: null,
                hasNextPage: false,
                hasPreviousPage: false,
                startCursor: null,
              },
              totalCount: deploymentsList.length || 0,
            });
          }
        });
      } else {
        reject(
          new GraphQLError(
            "Must provide one of: appIds, componentIds, installIds"
          )
        );
      }
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
