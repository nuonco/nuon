import { UpsertComponentRequest } from "@buf/nuon_apis.grpc_node/component/v1/messages_pb";
import { Config as BuildConfig } from "@buf/nuon_components.grpc_node/build/v1/build_pb";
import { Component } from "@buf/nuon_components.grpc_node/component/v1/component_pb";
import { Config as DeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/config_pb";
import { GraphQLError } from "graphql";
import type {
  BuildConfigInput,
  ComponentConfigInput,
  DeployConfigInput,
  Mutation,
  MutationUpsertComponentArgs,
  TResolverFn,
} from "../../../types";
import { formatComponent } from "../utils";
import { initBasicDeployConfig } from "./basic-deploy-config";
import { initDockerBuildConfig } from "./docker-build-config";
import { initExternalImageConfig } from "./external-image-config";
import { initHelmBuildConfig } from "./helm-build-config";
import { initHelmDeployConfig } from "./helm-deploy-config";
import { initNoopBuildConfig } from "./noop-build-config";
import { initNoopDeployConfig } from "./noop-deploy-config";
import { initTerraformBuildConfig } from "./terraform-build-config";
import { initTerraformDeployConfig } from "./terraform-deploy-config";

export function parseBuildConfigInput(
  buildConfig: BuildConfigInput
): Record<string, unknown> | any {
  const buildCfg = new BuildConfig();

  if (buildConfig?.noop) {
    const noopCfg = initNoopBuildConfig();

    buildCfg.setNoop(noopCfg);
  }

  if (buildConfig?.externalImageConfig) {
    const externalImageCfg = initExternalImageConfig(
      buildConfig?.externalImageConfig
    );

    buildCfg.setExternalImageCfg(externalImageCfg);
  }

  if (buildConfig?.dockerBuildConfig) {
    const dockerCfg = initDockerBuildConfig(buildConfig?.dockerBuildConfig);

    buildCfg.setDockerCfg(dockerCfg);
  }

  if (buildConfig?.helmBuildConfig) {
    const helmBuildCfg = initHelmBuildConfig(buildConfig?.helmBuildConfig);

    buildCfg.setHelmChartCfg(helmBuildCfg);
  }

  if (buildConfig?.terraformBuildConfig) {
    const terraformBuildCfg = initTerraformBuildConfig(
      buildConfig?.terraformBuildConfig
    );

    buildCfg.setTerraformModuleCfg(terraformBuildCfg);
  }

  return buildCfg;
}

export function parseDeployConfigInput(
  deployConfig: DeployConfigInput
): Record<string, unknown> | any {
  const deployCfg = new DeployConfig();

  if (deployConfig?.noop) {
    const noopCfg = initNoopDeployConfig();

    deployCfg.setNoop(noopCfg);
  }

  if (deployConfig?.basicDeployConfig) {
    const basicDeployCfg = initBasicDeployConfig(
      deployConfig?.basicDeployConfig
    );

    deployCfg.setBasic(basicDeployCfg);
  }

  if (deployConfig?.helmDeployConfig) {
    const helmDeployCfg = initHelmDeployConfig(deployConfig?.helmDeployConfig);

    deployCfg.setHelmChart(helmDeployCfg);
  }

  if (deployConfig?.terraformDeployConfig) {
    const terraformDeployCfg = initTerraformDeployConfig(
      deployConfig?.terraformDeployConfig
    );

    deployCfg.setTerraformModuleConfig(terraformDeployCfg);
  }

  return deployCfg;
}

export function parseConfigInput(
  config: ComponentConfigInput
): Record<string, unknown> | any {
  const componentConfig = new Component();

  if (config?.buildConfig) {
    componentConfig.setBuildCfg(parseBuildConfigInput(config?.buildConfig));
  }

  if (config?.deployConfig) {
    componentConfig.setDeployCfg(parseDeployConfigInput(config?.deployConfig));
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
