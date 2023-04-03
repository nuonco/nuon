import { UpsertComponentRequest } from "@buf/nuon_apis.grpc_node/component/v1/messages_pb";
import { Config as BuildConfig } from "@buf/nuon_components.grpc_node/build/v1/build_pb";
import { DockerConfig } from "@buf/nuon_components.grpc_node/build/v1/docker_pb";
import {
  AWSIAMAuthCfg,
  ExternalImageAuthConfig,
  ExternalImageConfig,
  PublicAuthCfg,
} from "@buf/nuon_components.grpc_node/build/v1/external_image_pb";
import { NoopConfig as NoopBuildConfig } from "@buf/nuon_components.grpc_node/build/v1/noop_pb";
import { Component } from "@buf/nuon_components.grpc_node/component/v1/component_pb";
import { BasicConfig as BasicDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/basic_pb";
import {
  Config as DeployConfig,
  ListenerConfig,
} from "@buf/nuon_components.grpc_node/deploy/v1/config_pb";
import { HelmRepoConfig as HelmRepoDeployConfig } from "@buf/nuon_components.grpc_node/deploy/v1/helm_repo_pb";
import { Config as VcsConfig } from "@buf/nuon_components.grpc_node/vcs/v1/config_pb";
import { ConnectedGithubConfig } from "@buf/nuon_components.grpc_node/vcs/v1/connected_github_pb";
import { PublicGitConfig } from "@buf/nuon_components.grpc_node/vcs/v1/public_git_pb";
import { GraphQLError } from "graphql";
import { TComponent, TResolverFn } from "../../types";
import { formatComponent } from "./utils";

interface IConnectedGithubConfigInput {
  branch?: string;
  directory: string;
  repo: string;
}
interface IPublicGitConfigInput {
  directory: string;
  repo: string;
}

interface IVcsConfigInput {
  connectedGithub?: IConnectedGithubConfigInput;
  publicGit?: IPublicGitConfigInput;
}

interface IDockerBuildConfigInput {
  dockerfile: string;
  vcsConfig: IVcsConfigInput;
}

interface IExternalImageInput {
  authConfig?: IExternalImageAuthConfig;
  ociImageUrl: string;
  tag?: string;
}

interface IExternalImageAuthConfig {
  region: "US_EAST_1" | "US_EAST_2" | "US_WEST_1" | "US_WEST_2";
  role: string;
}

interface IBuildConfigInput {
  dockerBuildConfig?: IDockerBuildConfigInput;
  externalImageConfig?: IExternalImageInput;
  noop?: boolean;
}

export function parseBuildConfigInput(
  buildConfig: IBuildConfigInput
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

  return buildCfg;
}

interface IBasicDeployConfigInput {
  healthCheckPath: string;
  instanceCount: number;
  port: number;
}

interface IHelmRepoDeployConfigInput {
  chartName: string;
  chartRepo: string;
  chartVersion: string;
  imageRepoValuesKey: string;
  imageTagValuesKey: string;
}

interface IDeployConfigInput {
  basicDeployConfig?: IBasicDeployConfigInput;
  helmRepoDeployConfig?: IHelmRepoDeployConfigInput;
  noop?: boolean;
}

export function parseDeployConfigInput(
  deployConfig: IDeployConfigInput
): Record<string, unknown> | any {
  const deployCfg = new DeployConfig();

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

  return deployCfg;
}

interface IComponentConfigInput {
  buildConfig?: IBuildConfigInput;
  deployConfig?: IDeployConfigInput;
}

export function parseConfigInput(
  config: IComponentConfigInput
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

interface IComponentInput {
  appId?: string;
  config?: IComponentConfigInput;
  id?: string;
  name?: string;
}

export const upsertComponent: TResolverFn<
  { input: IComponentInput },
  TComponent
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
