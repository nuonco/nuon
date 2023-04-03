import { TComponent } from "../../types";
import { getNodeFields } from "../../utils";

export function getVcsConfig(config) {
  let vcsConfig = {
    __typename: "NoopConfig",
  };
  const { connectedGithubConfig, publicGitConfig } = config?.vcsCfg;

  if (connectedGithubConfig) {
    vcsConfig = {
      __typename: "ConnectedGithubConfig",
      ...connectedGithubConfig,
    };
  } else if (publicGitConfig) {
    vcsConfig = {
      __typename: "PublicGitConfig",
      ...publicGitConfig,
    };
  }

  return vcsConfig;
}

export function getConfig(config) {
  let buildConfig = null;
  let deployConfig = null;

  if (config?.buildCfg) {
    if (config?.buildCfg?.noop) {
      buildConfig = {
        __typename: "NoopConfig",
        noop: true,
      };
    }

    if (config?.buildCfg?.externalImageCfg) {
      let authConfig = null;
      if (
        config?.buildCfg?.externalImageCfg?.authCfg?.awsIamAuthCfg?.awsRegion
      ) {
        authConfig = {
          __typename: "AWSAuthConfig",
          region:
            config?.buildCfg?.externalImageCfg?.authCfg?.awsIamAuthCfg
              ?.awsRegion,
          role: config?.buildCfg?.externalImageCfg?.authCfg?.awsIamAuthCfg
            ?.iamRoleArn,
        };
      }
      buildConfig = {
        __typename: "ExternalImageConfig",
        ...config?.buildCfg?.externalImageCfg,
        authConfig,
      };
    }

    if (config?.buildCfg?.dockerCfg) {
      const vcsConfig = getVcsConfig(config?.buildCfg?.dockerCfg);

      buildConfig = {
        __typename: "DockerBuildConfig",
        ...config?.buildCfg?.dockerCfg,
        vcsConfig,
      };
    }
  }

  if (config?.deployCfg) {
    if (config?.deployCfg?.basic) {
      const { basic } = config?.deployCfg;

      deployConfig = {
        __typename: "BasicDeployConfig",
        healthCheckPath: basic?.listenerCfg?.healthCheckPath,
        instanceCount: basic?.instanceCount,
        port: basic?.listenerCfg?.listenPort,
      };
    }

    if (config?.deployCfg?.helmRepo) {
      const { helmRepo } = config?.deployCfg;

      deployConfig = {
        __typename: "HelmRepoDeployConfig",
        ...helmRepo,
      };
    }
  }

  return {
    __typename: "ComponentConfig",
    buildConfig,
    deployConfig,
  };
}

export function formatComponent(component): TComponent {
  const config = getConfig(component.componentConfig);

  delete component.componentConfig;

  return {
    ...getNodeFields(component),
    config,
  };
}
