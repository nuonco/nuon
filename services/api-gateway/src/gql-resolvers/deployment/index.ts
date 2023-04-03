import { createDeployment } from "./create-deployment";
import { deployment } from "./deployment";
import { deploymentStatus } from "./deployment-status";
import { deployments } from "./deployments";
import { instance } from "./instance";

export const deploymentResolvers = {
  Mutation: {
    createDeployment,
  },
  Query: {
    deployment,
    deployments,
    deploymentStatus,
    instance,
  },
};
