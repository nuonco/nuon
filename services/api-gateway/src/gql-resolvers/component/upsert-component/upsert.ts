import { GraphQLError } from "graphql";
import { UpsertComponentRequest } from "../../../build/api/component/v1/messages_pb";
import { Component } from "../../../build/components/component/v1/component_pb";
import type {
  ComponentConfigInput,
  Mutation,
  MutationUpsertComponentArgs,
  TResolverFn,
  TgRPCMessage,
} from "../../../types";
import { formatComponent } from "../utils";
import { parseBuildConfigInput } from "./parse-build-config";
import { parseDeployConfigInput } from "./parse-deploy-config";

export function parseConfigInput(config: ComponentConfigInput): TgRPCMessage {
  const componentConfig = new Component();

  if (config?.buildConfig) {
    componentConfig.setBuildCfg(parseBuildConfigInput(config.buildConfig));
  }

  if (config?.deployConfig) {
    componentConfig.setDeployCfg(parseDeployConfigInput(config.deployConfig));
  }

  return componentConfig;
}

export const upsertComponent: TResolverFn<
  MutationUpsertComponentArgs,
  Mutation["upsertComponent"]
> = (_, { input }, { clients, user }) =>
  new Promise((resolve, reject) => {
    if (clients.component) {
      const request = new UpsertComponentRequest()
        .setAppId(input.appId)
        .setId(input.id)
        .setName(input.name)
        .setCreatedById(user?.id);

      if (input.config) {
        request.setComponentConfig(parseConfigInput(input.config));
      }

      clients.component.upsertComponent(request, (err, res) => {
        if (err) {
          reject(new GraphQLError(err.message));
        } else {
          resolve(formatComponent(res.toObject().component));
        }
      });
    } else {
      reject(new GraphQLError("Service isn't available"));
    }
  });
