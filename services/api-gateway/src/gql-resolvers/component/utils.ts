import type {
  Component,
  ComponentConfig,
  KeyValuePair,
  TerraformDeployConfig,
  VcsConfig,
} from "../../types";
import { getNodeFields } from "../../utils";

const ETerraformVersion = {
  0: "TERRAFORM_VERSION_UNSPECIFIED",
  1: "TERRAFORM_VERSION_LATEST",
};

export function getVcsConfig(config): VcsConfig {
  let vcsConfig = null;
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

export function getEnvVarsConfig(config): KeyValuePair {
  let envVarsConfig = null;

  if (config?.envVarsConfig) {
    envVarsConfig = {
      __typename: "KeyValuePair",
      ...config?.envVarsConfig,
    };
  }
  return envVarsConfig;
}

export function getConfig(config): ComponentConfig {
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
      const envVarsConfig = getEnvVarsConfig(config?.buildCfg?.dockerCfg);

      buildConfig = {
        __typename: "DockerBuildConfig",
        ...config?.buildCfg?.dockerCfg,
        envVarsConfig,
        vcsConfig,
      };
    }

    if (config?.buildCfg?.helmChartCfg) {
      const vcsConfig = getVcsConfig(config?.buildCfg?.helmChartCfg);
      const envVarsConfig = getEnvVarsConfig(config?.buildCfg?.helmChartCfg);

      delete config?.buildCfg?.helmChartCfg?.vcsCfg;

      buildConfig = {
        __typename: "HelmBuildConfig",
        ...config?.buildCfg?.helmChartCfg,
        envVarsConfig,
        vcsConfig,
      };
    }

    if (config?.buildCfg?.terraformModuleCfg) {
      const vcsConfig = getVcsConfig(config?.buildCfg?.terraformModuleCfg);
      delete config?.buildCfg?.terraformModuleCfg?.vcsCfg;

      buildConfig = {
        __typename: "TerraformBuildConfig",
        ...config?.buildCfg?.terraformModuleCfg,
        vcsConfig,
      };
    }
  }

  if (config?.deployCfg) {
    if (config?.deployCfg?.noop) {
      deployConfig = {
        __typename: "NoopConfig",
        noop: true,
      };
    }

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

    if (config?.deployCfg?.helmChart) {
      // TODO(nnnnat): temp until we know the fields for the helm chart deploy config
      deployConfig = {
        __typename: "HelmDeployConfig",
        noop: true,
        values:
          config?.deployCfg?.helmChart?.values?.valuesList?.map(
            ({ name, sensitive, value }) => ({
              __typename: "KeyValuePair",
              key: name,
              sensitive,
              value,
            })
          ) || null,
      };
    }

    if (config?.deployCfg?.terraformModuleConfig) {
      const terraformModuleConfig: TerraformDeployConfig = {
        terraformVersion:
          ETerraformVersion[
            config?.deployCfg?.terraformModuleConfig?.terraformVersion
          ],
        vars:
          config?.deployCfg?.terraformModuleConfig?.vars?.variablesList?.map(
            ({ name, sensitive, value }) => ({
              __typename: "KeyValuePair",
              key: name,
              sensitive,
              value,
            })
          ) || null,
      };

      deployConfig = {
        __typename: "TerraformDeployConfig",
        ...terraformModuleConfig,
      };
    }
  }

  return {
    __typename: "ComponentConfig",
    buildConfig,
    deployConfig,
  };
}

export function formatComponent(component): Component {
  const config = getConfig(component.componentConfig);

  delete component.componentConfig;

  return {
    ...getNodeFields(component),
    config,
  };
}
