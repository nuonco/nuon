import { build } from "../build/build";
import { component } from "../component/component";
import { deploy } from "../deploy/deploy";
import { instances } from "./instances";

export const instanceResolvers = {
  Instance: {
    build: (parent, _, ctx) => build(undefined, { id: parent.buildId }, ctx),
    component: (parent, _, ctx) =>
      component(undefined, { id: parent.componentId }, ctx),
    deploy: (parent, _, ctx) => deploy(undefined, { id: parent.deployId }, ctx),
  },
  Query: {
    instances,
  },
};
