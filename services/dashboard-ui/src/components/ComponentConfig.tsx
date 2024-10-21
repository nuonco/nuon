import React, { type FC } from 'react'
import { FaDocker, FaGitAlt, FaGithub } from 'react-icons/fa'
import { GoQuestion } from 'react-icons/go'
import {
  SiAwslambda,
  SiOpencontainersinitiative,
  SiHelm,
  SiTerraform,
} from 'react-icons/si'
import { Config, ConfigContent, Link, Text, ToolTip } from '@/components'
import { TComponentConfig } from '@/types'

export type TComponentConfigType =
  | 'docker'
  | 'external'
  | 'helm'
  | 'job'
  | 'terraform'

export function getComponentConfigType({
  docker_build,
  external_image,
  helm,
  job,
  terraform_module,
}: TComponentConfig): TComponentConfigType {
  return ((docker_build && 'docker') ||
    (external_image && 'external') ||
    (helm && 'helm') ||
    (terraform_module && 'terraform') ||
    (job && 'job')) as TComponentConfigType
}

export function getComponentConfigValues({
  docker_build,
  external_image,
  helm,
  job,
  terraform_module,
}: TComponentConfig): false | Record<string, any> {
  return (
    (!!docker_build && docker_build) ||
    (!!external_image && external_image) ||
    (!!helm && helm) ||
    (!!terraform_module && terraform_module) ||
    (!!job && job)
  )
}

export const StaticComponentConfigType: FC<{
  configType: string
  isIconOnly?: boolean
}> = ({ configType, isIconOnly = false }) => {
  let cfgType = {}
  switch (configType) {
    case 'docker':
      cfgType = { icon: <FaDocker />, name: 'Docker' }
      break
    case 'external':
      cfgType = { icon: <SiOpencontainersinitiative />, name: 'External image' }
      break
    case 'helm':
      cfgType = { icon: <SiHelm />, name: 'Helm' }
      break
    case 'terraform':
      cfgType = { icon: <SiTerraform />, name: 'Terraform' }
      break
    case 'job':
      cfgType = { icon: <SiAwslambda />, name: 'Job' }
      break
    default:
      cfgType = { icon: <GoQuestion />, name: 'Unknown' }
  }

  return (
    <span className="flex items-center gap-1">
      {cfgType['icon']} {!isIconOnly && cfgType['name']}
    </span>
  )
}

export interface IComponentConfiguration {
  config: TComponentConfig
}

export const ComponentConfiguration: FC<IComponentConfiguration> = ({
  config,
}) => {
  return (
    <div className="flex flex-col gap-8">
      <Config>
        <ConfigContent label="Version" value={config.version} />
        {config.terraform_module && <ConfigurationTerraform {...config} />}
        {config.docker_build && <ConfigurationDocker {...config} />}
        {config.external_image && <ConfigurationExternalImage {...config} />}
        {config.helm && <ConfigurationHelm {...config} />}
        {config.job && <ConfigurationJob {...config} />}
      </Config>

      {config.terraform_module && (
        <>
          {Object.keys(config.terraform_module?.variables).length !== 0 && (
            <ConfigurationVariables
              variables={config?.terraform_module?.variables}
            />
          )}
          {Object.keys(config.terraform_module?.env_vars).length !== 0 && (
            <ConfigurationVariables
              heading="Enviornment variables"
              variables={config?.terraform_module?.env_vars}
            />
          )}
        </>
      )}

      {config.docker_build && (
        <>
          {config.docker_build?.env_vars &&
            Object.keys(config.docker_build?.env_vars)?.length !== 0 && (
              <ConfigurationVariables
                heading="Enviornnment variables"
                variables={config?.docker_build?.env_vars}
              />
            )}
          {/* TODO(nnnnat): handle build args? */}
        </>
      )}

      {config.job && (
        <>
          {config.job?.env_vars &&
            Object.keys(config.job?.env_vars)?.length !== 0 && (
              <ConfigurationVariables
                heading="Enviornnment variables"
                variables={config?.job?.env_vars}
              />
            )}
          {/* TODO(nnnnat): handle args? */}
        </>
      )}

      {config.helm && (
        <>
          {config.helm?.values &&
            Object.keys(config?.helm?.values).length !== 0 && (
              <ConfigurationVariables
                heading="Config values"
                variables={config?.helm?.values}
              />
            )}
        </>
      )}
    </div>
  )
}

const ConfigurationTerraform: FC<
  Pick<TComponentConfig, 'terraform_module'>
> = ({ terraform_module }) => {
  return (
    <>
      <ConfigContent label="Terraform" value={terraform_module.version} />
      <ConfigurationVCS vcs={terraform_module} />
    </>
  )
}

const ConfigurationDocker: FC<Pick<TComponentConfig, 'docker_build'>> = ({
  docker_build,
}) => {
  return (
    <>
      <ConfigurationVCS vcs={docker_build} />
    </>
  )
}

const ConfigurationExternalImage: FC<
  Pick<TComponentConfig, 'external_image'>
> = ({ external_image }) => {
  return (
    <>
      <ConfigContent label="Image" value={external_image.image_url} />
      <ConfigContent label="Directory" value={external_image.tag} />
    </>
  )
}

const ConfigurationHelm: FC<Pick<TComponentConfig, 'helm'>> = ({ helm }) => {
  return (
    <>
      <ConfigurationVCS vcs={helm} />
    </>
  )
}

const ConfigurationJob: FC<Pick<TComponentConfig, 'job'>> = ({ job }) => {
  return (
    <>
      <ConfigContent label="Image" value={job.image_url} />
      <ConfigContent label="Tag" value={job.tag} />
      <ConfigContent label="Command" value={job.cmd} />
    </>
  )
}

const ConfigurationVCS: FC<{ vcs: any }> = ({ vcs }) => {
  const isGithubConnected = Boolean(vcs.connected_github_vcs_config)
  const repo = vcs.connected_github_vcs_config || vcs.public_git_vcs_config

  return (
    <>
      <ConfigContent
        label="Repository"
        value={
          <Link
            href={`https://github.com/${repo?.repo}`}
            target="_blank"
            rel="noreferrer"
          >
            {isGithubConnected ? <FaGithub /> : <FaGitAlt />}
            {repo?.repo}
          </Link>
        }
      />

      <ConfigContent label="Directory" value={repo?.directory} />
      <ConfigContent label="Branch" value={repo?.branch} />
    </>
  )
}

const ConfigurationVariables: FC<{
  heading?: string
  variables: Record<string, string>
}> = ({ heading = 'Variables', variables }) => {
  const variableKeys = Object.keys(variables)
  const isEmpty = variableKeys.length === 0

  return (
    !isEmpty && (
      <div className="flex flex-col gap-4">
        <div className="">
          <Text className="text-sm !font-medium leading-normal">{heading}</Text>
        </div>

        <div className="divide-y">
          <div className="grid grid-cols-3 gap-4 pb-3">
            <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
              Name
            </Text>
            <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
              Value
            </Text>
          </div>

          {variableKeys.map((key, i) => (
            <div key={`${key}-${i}`} className="grid grid-cols-3 gap-4 py-3">
              <Text className="font-mono text-sm break-all">{key}</Text>
              <ToolTip tipContent={variables[key]} alignment="right">
                <Text className="font-mono text-sm break-all col-span-2 !inline truncate max-w-[200px]">
                  {variables[key]}
                </Text>
              </ToolTip>
            </div>
          ))}
        </div>
      </div>
    )
  )
}

const ConfigurationArguments: FC<{ args: Array<string> }> = () => {
  return <>Args</>
}
