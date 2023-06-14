import { component } from "../component/component";
import { deploy } from "../deploy/deploy";
import { instances } from "./instances";

export const instanceResolvers = {
  Instance: {
    component: (parent, _, ctx) =>
      component(undefined, { id: parent.componentId }, ctx),
    deploy: (parent, _, ctx) =>
      deploy(undefined, { instanceId: parent.id }, ctx),
  },
  Query: {
    instances,
  },
};
