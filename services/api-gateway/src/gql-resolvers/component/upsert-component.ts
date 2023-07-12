import { UpsertComponentRequest } from "@buf/nuon_apis.grpc_node/component/v1/messages_pb";
import { Config as BuildConfig } from "@buf/nuon_components.grpc_node/build/v1/build_pb";
import { DockerConfig } from "@buf/nuon_components.grpc_node/build/v1/docker_pb";
import {
  AWSIAMAuthCfg,
  ExternalImageAuthConfig,
  ExternalImageConfig,
  PublicAuthCfg,
} from "@buf/nuon_components.grpc_node/build/v1/external_image_pb";
import { HelmChartConfig } from "@buf/nuon_components.grpc_node/build/v1/helm_chart_pb";
import { NoopConfig as NoopBuildConfig } from "@buf/nuon_components.grpc_node/build/v1/noop_pb";
import { TerraformModuleConfig as TerraformBuildConfig } from "@buf/nuon_components.grpc_node/build/v1/terraform_module_pb";
import { Component } from "@buf/nuon_components.grpc_node/component/v1/component_pb";
import { BasicConfig as BasicDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/basic_pb";
import {
  Config as DeployConfig,
  ListenerConfig,
} from "@buf/nuon_components.grpc_node/deploy/v1/config_pb";
import { HelmChartConfig as HelmDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/helm_chart_pb";
import { HelmRepoConfig as HelmRepoDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/helm_repo_pb";
import { NoopConfig as NoopDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/noop_pb";
import { TerraformModuleConfig as TerraformDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/terraform_module_pb";
import {
  HelmValue,
  HelmValues,
} from "@buf/nuon_components.grpc_node/variables/v1/helm_pb";
import { Config as VcsConfig } from "@buf/nuon_components.grpc_node/vcs/v1/config_pb";
import { ConnectedGithubConfig } from "@buf/nuon_components.grpc_node/vcs/v1/connected_github_pb";
import { PublicGitConfig } from "@buf/nuon_components.grpc_node/vcs/v1/public_git_pb";
import { GraphQLError } from "graphql";
import type {
  BuildConfigInput,
  ComponentConfigInput,
  DeployConfigInput,
  Mutation,
  MutationUpsertComponentArgs,
  TResolverFn,
} from "../../types";
import { formatComponent } from "./utils";

export function parseBuildConfigInput(
  buildConfig: BuildConfigInput
): Record<string, unknown> | any {
  const buildCfg = new BuildConfig();

  if (buildConfig?.noop) {
    const noopCfg = new NoopBuildConfig();

    buildCfg.setNoop(noopCfg);
  }

  if (buildConfig?.externalImageConfig) {
    const { authConfig, ociImageUrl, tag } = buildConfig?.externalImageConfig;

    const externalImageAuthCfg = new ExternalImageAuthConfig();
    if (authConfig) {
      const privateAuthCfg = new AWSIAMAuthCfg()
        .setIamRoleArn(authConfig.role)
        .setAwsRegion(authConfig.region);
      externalImageAuthCfg.setAwsIamAuthCfg(privateAuthCfg);
    } else {
      const publicAuthCfg = new PublicAuthCfg();
      externalImageAuthCfg.setPublicAuthCfg(publicAuthCfg);
    }

    const externalImageCfg = new ExternalImageConfig()
      .setOciImageUrl(ociImageUrl)
      .setTag(tag)
      .setAuthCfg(externalImageAuthCfg);

    buildCfg.setExternalImageCfg(externalImageCfg);
  }

  if (buildConfig?.dockerBuildConfig) {
    const { dockerfile, vcsConfig: vcsInput } = buildConfig?.dockerBuildConfig;
    const vcsConfig = new VcsConfig();

    if (vcsInput?.connectedGithub) {
      const connectedGithubCfg = new ConnectedGithubConfig()
        .setRepo(vcsInput?.connectedGithub?.repo)
        .setDirectory(vcsInput?.connectedGithub?.directory);

      vcsConfig.setConnectedGithubConfig(connectedGithubCfg);
    } else if (vcsInput?.publicGit) {
      const publicGitCfg = new PublicGitConfig()
        .setRepo(vcsInput?.publicGit?.repo)
        .setDirectory(vcsInput?.publicGit?.directory);

      vcsConfig.setPublicGitConfig(publicGitCfg);
    }

    const dockerCfg = new DockerConfig()
      .setDockerfile(dockerfile)
      .setVcsCfg(vcsConfig);

    buildCfg.setDockerCfg(dockerCfg);
  }

  if (buildConfig?.helmBuildConfig) {
    const { chartName, vcsConfig: vcsInput } = buildConfig?.helmBuildConfig;
    const vcsConfig = new VcsConfig();

    if (vcsInput?.connectedGithub) {
      const connectedGithubCfg = new ConnectedGithubConfig()
        .setRepo(vcsInput?.connectedGithub?.repo)
        .setDirectory(vcsInput?.connectedGithub?.directory);

      vcsConfig.setConnectedGithubConfig(connectedGithubCfg);
    } else if (vcsInput?.publicGit) {
      const publicGitCfg = new PublicGitConfig()
        .setRepo(vcsInput?.publicGit?.repo)
        .setDirectory(vcsInput?.publicGit?.directory);

      vcsConfig.setPublicGitConfig(publicGitCfg);
    }

    const helmBuildConfig = new HelmChartConfig()
      .setChartName(chartName)
      .setVcsCfg(vcsConfig);

    buildCfg.setHelmChartCfg(helmBuildConfig);
  }

  if (buildConfig?.terraformBuildConfig) {
    const { vcsConfig: vcsInput } = buildConfig?.terraformBuildConfig;
    const vcsConfig = new VcsConfig();

    if (vcsInput?.connectedGithub) {
      const connectedGithubCfg = new ConnectedGithubConfig()
        .setRepo(vcsInput?.connectedGithub?.repo)
        .setDirectory(vcsInput?.connectedGithub?.directory);

      vcsConfig.setConnectedGithubConfig(connectedGithubCfg);
    } else if (vcsInput?.publicGit) {
      const publicGitCfg = new PublicGitConfig()
        .setRepo(vcsInput?.publicGit?.repo)
        .setDirectory(vcsInput?.publicGit?.directory);

      vcsConfig.setPublicGitConfig(publicGitCfg);
    }

    const terraformBuildConfig = new TerraformBuildConfig().setVcsCfg(
      vcsConfig
    );

    buildCfg.setTerraformModuleCfg(terraformBuildConfig);
  }

  return buildCfg;
}

export function parseDeployConfigInput(
  deployConfig: DeployConfigInput
): Record<string, unknown> | any {
  const deployCfg = new DeployConfig();

  if (deployConfig?.noop) {
    const noopCfg = new NoopDeployConfig();

    deployCfg.setNoop(noopCfg);
  }

  if (deployConfig?.basicDeployConfig) {
    const { healthCheckPath, instanceCount, port } =
      deployConfig?.basicDeployConfig;
    const listenerCfg = new ListenerConfig()
      .setListenPort(port)
      .setHealthCheckPath(healthCheckPath);
    const basicDeployCfg = new BasicDeployConfig()
      .setInstanceCount(instanceCount)
      .setListenerCfg(listenerCfg);

    deployCfg.setBasic(basicDeployCfg);
  }

  if (deployConfig?.helmDeployConfig) {
    // TODO(nnnnat): temp until we know the required fields
    const helmDeployCfg = new HelmDeployConfig();

    if (deployConfig?.helmDeployConfig?.values) {
      const { values } = deployConfig?.helmDeployConfig;
      const helmValues = new HelmValues().setValuesList(
        values.map(({ key, sensitive, value }) => {
          return new HelmValue()
            .setName(key)
            .setValue(value)
            .setSensitive(sensitive);
        })
      );

      helmDeployCfg.setValues(helmValues);
    }

    deployCfg.setHelmChart(helmDeployCfg);
  }

  if (deployConfig?.helmRepoDeployConfig) {
    const {
      chartName,
      chartRepo,
      chartVersion,
      imageRepoValuesKey,
      imageTagValuesKey,
    } = deployConfig?.helmRepoDeployConfig;
    const helmRepoDeployCfg = new HelmRepoDeployConfig()
      .setChartName(chartName)
      .setChartRepo(chartRepo)
      .setChartVersion(chartVersion)
      .setImageRepoValuesKey(imageRepoValuesKey)
      .setImageTagValuesKey(imageTagValuesKey);

    deployCfg.setHelmRepo(helmRepoDeployCfg);
  }

  if (deployConfig?.terraformDeployConfig) {
    const terraformDeployCfg = new TerraformDeployConfig().setTerraformVersion(
      1 //deployConfig?.terraformDeployConfig
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
