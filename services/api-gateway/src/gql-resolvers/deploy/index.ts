import { deploy } from "./deploy";
import { deployLog } from "./deploy-log";
import { startDeploy } from "./start-deploy";

export const deployResolvers = {
  Mutation: {
    startDeploy,
  },
  Query: {
    deploy,
    deployLog,
  },
};
