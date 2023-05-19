import { deploy } from "./deploy";
import { startDeploy } from "./start-deploy";

export const deployResolvers = {
  Mutation: {
    startDeploy,
  },
  Query: {
    deploy,
  },
};
