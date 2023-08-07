import { build } from "../build/build";
import { install } from "../install/install";
import { deploy } from "./deploy";
import { startDeploy } from "./start-deploy";

export const deployResolvers = {
  Deploy: {
    build: (parent, _, ctx) => build(parent, { id: parent.buildId }, ctx),
    install: (parent, _, ctx) => install(parent, { id: parent.installId }, ctx),
  },
  Mutation: {
    startDeploy,
  },
  Query: {
    deploy,
  },
};
