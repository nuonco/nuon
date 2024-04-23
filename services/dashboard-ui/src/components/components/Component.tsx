import React, { type FC, } from 'react'
import { FaDocker} from 'react-icons/fa'
import {
  GoQuestion,
} from 'react-icons/go'
import {
  SiAwslambda,
  SiOpencontainersinitiative,
  SiHelm,
  SiTerraform,
} from 'react-icons/si'
import { Code, Heading, Text, VCS } from '@/components'
import { TComponentConfig } from '@/types'

export type TComponentType =
  | 'docker'
  | 'external'
  | 'helm'
  | 'job'
  | 'terraform'

export function getComponentTypeFromConfig({
  docker_build,
  external_image,
  helm,
  job,
  terraform_module,
}: TComponentConfig): TComponentType {
  return ((docker_build && 'docker') ||
    (external_image && 'external') ||
    (helm && 'helm') ||
    (terraform_module && 'terraform') ||
    (job && 'job')) as TComponentType
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

export const ComponentType: FC<{
  config?: TComponentConfig
  componentType?: TComponentType
  hasTextHidden?: boolean
}> = ({ config, componentType, hasTextHidden }) => {
  let ct =
    componentType || getComponentTypeFromConfig(config as TComponentConfig)
  let el: React.ReactNode

  switch (ct) {
    case 'docker':
      el = (
        <>
          <FaDocker /> {!hasTextHidden && 'Docker'}
        </>
      )
      break

    case 'external':
      el = (
        <>
          <SiOpencontainersinitiative /> {!hasTextHidden && 'External image'}
        </>
      )
      break
    case 'helm':
      el = (
        <>
          <SiHelm /> {!hasTextHidden && 'Helm'}
        </>
      )
      break

    case 'terraform':
      el = (
        <>
          <SiTerraform /> {!hasTextHidden && 'Terraform'}
        </>
      )
      break

    case 'job':
      el = (
        <>
          <SiAwslambda /> {!hasTextHidden && 'Job'}
        </>
      )
      break

    default:
      el = (
        <>
          <GoQuestion /> {!hasTextHidden && 'Unkown'}
        </>
      )
  }

  return <span className="flex items-center gap-1">{el}</span>
}

export const ComponentDependencies: FC<{ deps: Array<string> }> = ({
  deps,
}) => {
  return (
    <div className="flex flex-col gap-4">
      {deps.map((d) => (
        <span key={d}>{d}</span>
      ))}
    </div>
  )
}

export const ComponentConfig: FC<{
  config: TComponentConfig
  version?: number
}> = ({ config, version }) => {
  const componentType = getComponentTypeFromConfig(config)

  return (
    <div className="flex flex-col gap-2">
      <ComponentType config={config} />
      {(componentType === 'docker' && (
        <DockerConfig {...{ cfg: config?.docker_build, version }} />
      )) ||
        (componentType === 'terraform' && (
          <TerraformConfig {...{ cfg: config?.terraform_module, version }} />
        ))}
    </div>
  )
}

export const DockerConfig: FC<{
  cfg: TComponentConfig['docker_build']
  version?: number
}> = ({ cfg, version }) => {
  return (
    <>
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <VCS {...cfg} />
      </span>

      <Heading variant="subheading">Build arguments</Heading>
      <Code variant="preformated">
        {JSON.stringify(cfg?.build_args, null, 2)}
      </Code>

      {cfg?.env_vars !== null && (
        <>
          <Heading variant="subheading">Enviroment variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.env_vars, null, 2)}
          </Code>
        </>
      )}
    </>
  )
}

export const TerraformConfig: FC<{
  cfg: TComponentConfig['terraform_module']
  version?: number
}> = ({ cfg, version }) => {
  return (
    <>
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <Text variant="caption">
          <b>Terraform version:</b> {cfg?.version as string}
        </Text>
        <VCS {...cfg} />
      </span>

      <Heading variant="subheading">Terraform veriables</Heading>
      <Code variant="preformated">
        {JSON.stringify(cfg?.variables, null, 2)}
      </Code>

      {cfg?.env_vars !== null && (
        <>
          <Heading variant="subheading">Enviroment variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.env_vars, null, 2)}
          </Code>
        </>
      )}
    </>
  )
}
