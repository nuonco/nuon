// @ts-nocheck
'use client'

import classNames from 'classnames'
import React, { type FC, useState } from 'react'
import { createPortal } from 'react-dom'
import { FaDocker, FaGitAlt, FaGithub } from 'react-icons/fa'
import { GoQuestion } from 'react-icons/go'
import {
  SiAwslambda,
  SiHelm,
  SiKubernetes,
  SiOpencontainersinitiative,
  SiTerraform,
} from 'react-icons/si'
import { ArrowsOutSimple } from '@phosphor-icons/react/dist/ssr'
import { Button } from '@/components/Button'
import { Config, ConfigContent } from '@/components/Config'
import { CodeViewer } from '@/components/Code'
import { Modal } from '@/components/Modal'
import { Link } from '@/components/Link'
import { ToolTip } from '@/components/ToolTip'
import { Text, Truncate, type TTextVariant } from '@/components/Typography'
import { TComponentConfig, TVCSGit, TVCSGitHub } from '@/types'

type TComponentConfigType =
  | 'docker'
  | 'external'
  | 'helm'
  | 'job'
  | 'terraform'
  | 'kubernetes_manifest'

export function getComponentConfigType({
  docker_build,
  external_image,
  helm,
  job,
  terraform_module,
  kubernetes_manifest,
}: TComponentConfig): TComponentConfigType {
  return ((docker_build && 'docker') ||
    (external_image && 'external') ||
    (helm && 'helm') ||
    (terraform_module && 'terraform') ||
    (job && 'job') ||
    (kubernetes_manifest as 'kubernetes_manifest')) as TComponentConfigType
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
    case 'kubernetes_manifest':
      cfgType = { icon: <SiKubernetes />, name: 'Kubernetes Manifest' }
      break
    default:
      cfgType = { icon: <GoQuestion />, name: 'Unknown' }
  }

  return (
    <span className="flex items-center gap-1 text-nowrap">
      {cfgType['icon']} {!isIconOnly && cfgType['name']}
    </span>
  )
}

export interface IComponentConfiguration {
  config: TComponentConfig
  isNotTruncated?: boolean
  hideHelmValuesFile?: boolean
}

export const ComponentConfiguration: FC<IComponentConfiguration> = ({
  config,
  isNotTruncated = false,
  hideHelmValuesFile = false,
}) => {
  return (
    <div className="flex flex-col gap-8">
      <Config>
        <ConfigContent label="Version" value={config.version} />
        {config.terraform_module && <ConfigurationTerraform {...config} />}
        {config.docker_build && <ConfigurationDocker {...config} />}
        {config.external_image && <ConfigurationExternalImage {...config} />}
        {config.helm && <ConfigurationHelm {...config} />}
        {config.kubernetes_manifest && <ConfigurationK8s {...config} />}
        {config.job && <ConfigurationJob {...config} />}
      </Config>

      {config.terraform_module && (
        <>
          {Object.keys(config.terraform_module?.variables).length !== 0 && (
            <ConfigurationVariables
              variables={config?.terraform_module?.variables}
              isNotTruncated={isNotTruncated}
            />
          )}
          {Object.keys(config.terraform_module?.env_vars).length !== 0 && (
            <ConfigurationVariables
              heading="Enviornment variables"
              variables={config?.terraform_module?.env_vars}
              isNotTruncated={isNotTruncated}
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
                isNotTruncated={isNotTruncated}
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
                isNotTruncated={isNotTruncated}
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
                isNotTruncated={isNotTruncated}
              />
            )}
          {config?.helm?.values_files?.length && !hideHelmValuesFile ? (
            <div className="flex flex-col gap-4">
              <Text variant="med-12">Values file</Text>
              <CodeViewer
                initCodeSource={config?.helm?.values_files}
                language="yaml"
              />
            </div>
          ) : null}
        </>
      )}

      {config?.kubernetes_manifest && (
        <div className="flex flex-col gap-4">
          <Text variant="med-12">Manifest</Text>
          <CodeViewer
            initCodeSource={config?.kubernetes_manifest?.manifest}
            language="yaml"
          />
        </div>
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
      <ConfigContent
        label="Image"
        value={
          external_image.image_url?.length >= 32 ? (
            <ToolTip alignment="right" tipContent={external_image.image_url}>
              <Truncate variant="small">{external_image.image_url}</Truncate>
            </ToolTip>
          ) : (
            external_image.image_url
          )
        }
      />
      <ConfigContent
        label="Tag"
        value={
          external_image.tag.length >= 12 ? (
            <ToolTip alignment="right" tipContent={external_image.tag}>
              <Truncate variant="small">{external_image.tag}</Truncate>
            </ToolTip>
          ) : (
            external_image.tag
          )
        }
      />
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

const ConfigurationK8s: FC<Pick<TComponentConfig, 'kubernetes_manifest'>> = ({
  kubernetes_manifest,
}) => {
  return (
    <>
      <ConfigContent
        label="Namespace"
        value={
          kubernetes_manifest.namespace.length >= 32 ? (
            <ToolTip
              alignment="right"
              tipContent={kubernetes_manifest.namespace}
            >
              <Truncate variant="small">
                {kubernetes_manifest.namespace}
              </Truncate>
            </ToolTip>
          ) : (
            kubernetes_manifest.namespace
          )
        }
      />
    </>
  )
}

const ConfigurationJob: FC<Pick<TComponentConfig, 'job'>> = ({ job }) => {
  return (
    <>
      <ConfigContent
        label="Image"
        value={
          job.image_url.length >= 32 ? (
            <ToolTip alignment="right" tipContent={job.image_url}>
              <Truncate variant="small">{job.image_url}</Truncate>
            </ToolTip>
          ) : (
            job.image_url
          )
        }
      />
      <ConfigContent
        label="Tag"
        value={
          job.tag.length >= 12 ? (
            <ToolTip alignment="right" tipContent={job.tag}>
              <Truncate variant="small">{job.tag}</Truncate>
            </ToolTip>
          ) : (
            job.tag
          )
        }
      />
      <ConfigContent label="Command" value={job.cmd} />
    </>
  )
}

interface IConfigurationVCS {
  connected_github_vcs_config?: TVCSGitHub
  public_git_vcs_config?: TVCSGit
}

export const ConfigurationVCS: FC<{ vcs: IConfigurationVCS }> = ({ vcs }) => {
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
            {repo?.repo?.length >= 32 ? (
              <ToolTip alignment="right" tipContent={repo?.repo}>
                <Truncate variant="large">{repo?.repo}</Truncate>
              </ToolTip>
            ) : (
              repo?.repo
            )}
          </Link>
        }
      />

      <ConfigContent
        label="Directory"
        value={
          repo?.directory?.length >= 12 ? (
            <ToolTip alignment="right" tipContent={repo?.directory}>
              <Truncate variant="small">{repo?.directory}</Truncate>
            </ToolTip>
          ) : (
            repo?.directory
          )
        }
      />
      <ConfigContent
        label="Branch"
        value={
          repo?.branch?.length >= 12 ? (
            <ToolTip alignment="right" tipContent={repo?.branch}>
              <Truncate variant="small">{repo?.branch}</Truncate>
            </ToolTip>
          ) : (
            repo?.branch
          )
        }
      />
    </>
  )
}

export const ConfigurationVariables: FC<{
  headingVariant?: TTextVeriant
  heading?: string
  isNotTruncated?: boolean
  variables: Record<string, string>
}> = ({
  heading = 'Variables',
  headingVariant = 'med-12',
  isNotTruncated = false,
  variables,
}) => {
  const variableKeys = Object.keys(variables)
  const isEmpty = variableKeys.length === 0
  const [isOpen, setIsOpen] = useState(false)

  return (
    !isEmpty && (
      <>
        {isOpen
          ? createPortal(
              <Modal
                heading={heading}
                isOpen={isOpen}
                onClose={() => {
                  setIsOpen(false)
                }}
              >
                <ConfigVariables
                  keys={variableKeys}
                  variables={variables}
                  isNotTruncated
                />
              </Modal>,
              document.body
            )
          : null}
        <div className="flex flex-col gap-4">
          <div className="flex items-center justify-between">
            <Text variant={headingVariant}>{heading}</Text>

            <Button
              className="text-sm !font-medium flex items-center gap-2 !p-1"
              onClick={() => {
                setIsOpen(true)
              }}
              title={`Expand ${heading}`}
              variant="ghost"
            >
              <ArrowsOutSimple />
            </Button>
          </div>

          <ConfigVariables
            keys={variableKeys}
            variables={variables}
            isNotTruncated={isNotTruncated}
          />
        </div>
      </>
    )
  )
}

// TODO(nnnat): refactor this mess
export const ConfigVariables: FC<{
  keys: Array<string>
  variables: Record<string, string | string[]>
  isNotTruncated?: boolean
}> = ({ keys, variables, isNotTruncated = false }) => {
  return (
    <div className="grid grid-cols-[fit-content(30rem)_auto]">
      <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500 pb-3 pr-4 border-b">
        Name
      </Text>
      <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500 pb-3 pl-4 border-b">
        Value
      </Text>

      {keys.map((key, i) => (
        <React.Fragment key={`${key}-${i}`}>
          <Text
            className={classNames(
              'font-mono py-3 pr-4 content-baseline break-all',
              {
                'border-b': i + 1 !== keys?.length,
              }
            )}
          >
            {key.length >= 15 && !isNotTruncated ? (
              <ToolTip tipContent={key} alignment="left">
                <Truncate variant="small">{key}</Truncate>
              </ToolTip>
            ) : (
              key
            )}
          </Text>
          <Text
            className={classNames(
              'font-mono py-3 pl-4 content-baseline break-all',
              {
                'border-b': i + 1 !== keys?.length,
              }
            )}
          >
            {variables[key]?.length >= 24 && !isNotTruncated ? (
              <ToolTip tipContent={variables[key]} alignment="right">
                <Truncate variant="large">
                  {typeof variables[key] === 'string'
                    ? variables[key]
                    : variables[key]?.map((v) => <span key={key}>{v}</span>)}
                </Truncate>
              </ToolTip>
            ) : typeof variables[key] === 'string' ? (
              variables[key] ? (
                variables[key]
              ) : (
                ''
              )
            ) : Array.isArray(variables[key]) ? (
              variables[key]?.map((v, i) => (
                <span key={`${key}-${i}`}>
                  {v}
                  {i + 1 !== variables[key]?.length && ','}
                </span>
              ))
            ) : (
              <div className="flex flex-col gap-1 overflow-x-auto">
                {variables[key] ? (
                  <CodeViewer
                    initCodeSource={JSON.stringify(variables[key], null, 2)}
                  />
                ) : null}
              </div>
            )}
          </Text>
        </React.Fragment>
      ))}
    </div>
  )
}

const ConfigurationArguments: FC<{ args: Array<string> }> = () => {
  return <>Args</>
}
