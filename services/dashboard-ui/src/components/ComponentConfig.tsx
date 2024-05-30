import React, { Suspense, type FC } from 'react'
import { FaDocker } from 'react-icons/fa'
import { GoQuestion } from 'react-icons/go'
import {
  SiAwslambda,
  SiOpencontainersinitiative,
  SiHelm,
  SiTerraform,
} from 'react-icons/si'
import { Card, Code, Heading, Text, VCS } from '@/components'
import { getComponentConfig, type IGetComponentConfig } from '@/lib'
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

export const ComponentConfigType: FC<
  IGetComponentConfig & {
    isIconOnly?: boolean
  }
> = async ({ componentId, componentConfigId, orgId, isIconOnly = false }) => {
  let config: TComponentConfig
  try {
    config = await getComponentConfig({ orgId, componentId, componentConfigId })
  } catch (error) {
    return <>No config</>
  }

  let cfgType = {}
  switch (getComponentConfigType(config)) {
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

export interface IComponentConfig extends IGetComponentConfig {
  version?: number
}

export const ComponentConfig: FC<IComponentConfig> = async ({
  componentId,
  componentConfigId,
  orgId,
  version,
}) => {
  let config: TComponentConfig
  try {
    config = await getComponentConfig({ orgId, componentId, componentConfigId })
  } catch (error) {
    return <>No config found</>
  }

  const configType = getComponentConfigType(config)

  return (
    <div className="flex flex-col gap-2">
      {(configType === 'docker' && (
        <ComponentConfigDocker {...{ cfg: config?.docker_build, version }} />
      )) ||
        (configType === 'terraform' && (
          <ComponentConfigTerraform
            {...{ cfg: config?.terraform_module, version }}
          />
        )) ||
        (configType === 'job' && (
          <ComponentConfigJob {...{ cfg: config?.job, version }} />
        )) ||
        (configType === 'external' && (
          <ComponentConfigOCI {...{ cfg: config?.external_image, version }} />
        )) ||
        (configType === 'helm' && (
          <ComponentConfigHelm {...{ cfg: config?.helm, version }} />
        ))}
    </div>
  )
}

export const ComponentConfigCard: FC<
  IComponentConfig & { heading?: string }
> = ({ heading = 'Component config', ...props }) => (
  <Card>
    <Heading>{heading}</Heading>
    <Suspense fallback="Loading component config...">
      <ComponentConfig {...props} />
    </Suspense>
  </Card>
)

export const ComponentConfigDocker: FC<{
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

export const ComponentConfigTerraform: FC<{
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

      <Heading variant="subheading">Terraform variables</Heading>
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

export const ComponentConfigHelm: FC<{
  cfg: TComponentConfig['helm']
  version?: number
}> = ({ cfg, version }) => {
  return (
    <>
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <Text variant="caption">
          <b>Chart:</b> {cfg?.chart_name}
        </Text>
        <VCS {...cfg} />
      </span>

      {cfg?.values ? (
        <>
          <Heading variant="subheading">Config values</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.values, null, 4)}
          </Code>
        </>
      ) : null}
    </>
  )
}

export const ComponentConfigOCI: FC<{
  cfg: TComponentConfig['external_image']
  version?: number
}> = ({ cfg, version }) => {
  return (
    <>
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <Text variant="caption">
          <b>Image:</b> {cfg?.image_url}
        </Text>
        <Text variant="caption">
          <b>Tag:</b> {cfg?.tag}
        </Text>
      </span>

      {cfg?.env_vars ? (
        <>
          <Heading variant="subheading">Enviornment variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.env_vars, null, 4)}
          </Code>
        </>
      ) : null}
    </>
  )
}

export const ComponentConfigJob: FC<{
  cfg: TComponentConfig['job']
  version?: number
}> = ({ cfg, version }) => {
  return (
    <>
      <span>
        <Text variant="caption">
          <b>Config version:</b> {version}
        </Text>
        <Text variant="caption">
          <b>Image:</b> {cfg?.image_url}
        </Text>
        <Text variant="caption">
          <b>Tag:</b> {cfg?.tag}
        </Text>
        <Text variant="caption">
          <b>Command:</b> <span className="tx-code tx-small">{cfg?.cmd}</span>
        </Text>
      </span>

      {cfg?.args ? (
        <>
          <Heading variant="subheading">Arguments</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.args, null, 4)}
          </Code>
        </>
      ) : null}

      {cfg?.env_vars ? (
        <>
          <Heading variant="subheading">Enviornment variables</Heading>
          <Code variant="preformated">
            {JSON.stringify(cfg?.env_vars, null, 4)}
          </Code>
        </>
      ) : null}
    </>
  )
}
