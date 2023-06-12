import { createDeployment } from "./create-deployment";
import { deployment } from "./deployment";
import { deploymentStatus } from "./deployment-status";
import { deployments } from "./deployments";
import { instanceStatus } from "./instance-status";

export const deploymentResolvers = {
  Mutation: {
    createDeployment,
  },
  Query: {
    deployment,
    deployments,
    deploymentStatus,
    instanceStatus,
  },
};
