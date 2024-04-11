import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import { FaDocker, FaGitAlt, FaGithub } from 'react-icons/fa'
import {
  GoArrowLeft,
  GoKebabHorizontal,
  GoCheckCircleFill,
  GoClockFill,
  GoContainer,
  GoXCircleFill,
  GoInfo,
  GoQuestion,
} from 'react-icons/go'
import {
  SiAwslambda,
  SiOpencontainersinitiative,
  SiHelm,
  SiTerraform,
} from 'react-icons/si'
import {
  Card,
  Code,
  StatusTimeline,
  Heading,
  Link,
  Status,
  Text,
  VCS,
} from '@/components'
import type { TComponent, TComponentConfig, TInstallComponent } from '@/types'

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
  return (
    (docker_build && 'docker') ||
    (external_image && 'external') ||
    (helm && 'helm') ||
    (terraform_module && 'terraform') ||
    (job && 'job')
  )
}

export function getComponentConfigValues({
  docker_build,
  external_image,
  helm,
  job,
  terraform_module,
}: TComponentConfig): Record<string, unknown> {
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
  let ct = componentType || getComponentTypeFromConfig(config)
  let el

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
    (componentType === 'docker' && (
      <DockerConfig {...{ cfg: config?.docker_build, version }} />
    )) ||
    (componentType === 'terraform' && (
      <TerraformConfig {...{ cfg: config?.terraform_module, version }} />
    ))
  )
}

export const DockerConfig: FC<{
  cfg: TComponentConfig['docker_build']
  version: number
}> = ({ cfg, version }) => {
  return (
    <div className="flex flex-col gap-2">
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
    </div>
  )
}

export const TerraformConfig: FC<{
  cfg: TComponentConfig['terraform_module']
  version: number
}> = ({ cfg, version }) => {
  return (
    <div className="flex flex-col gap-2">
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <Text variant="caption">
          <b>Terraform version:</b> {cfg?.version}
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
    </div>
  )
}
