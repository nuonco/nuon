'use client'

import { DateTime } from 'luxon'
import React, { type FC, useEffect, useState } from 'react'
import { FaAws, FaGitAlt, FaGithub } from 'react-icons/fa'
import {
  GoArrowLeft,
  GoKebabHorizontal,
  GoCheckCircleFill,
  GoClockFill,
  GoXCircleFill,
  GoInfo,
} from 'react-icons/go'
import { VscAzure } from 'react-icons/vsc'
import { Card, Code, Heading, Link, Status, Text } from '@/components'
import type {
  TInstall,
  TInstallAwsAccount,
  TInstallAzureAccount,
  TInstallComponent,
  TSandboxConfig,
} from '@/types'
import {
  AWS_REGIONS,
  AZURE_REGIONS,
  getFlagEmoji,
  getFullInstallStatus,
} from '@/utils'

export const Install: FC<{ install: TInstall; orgId: string }> = ({
  install,
}) => {
  return (
    <div className="flex flex-col xl:flex-row flex-auto flex-wrap gap-4 items-start justify-between">
      <div className="flex flex-col gap-1 flex-auto">
        <InstallStatus install={install} isCompact />
        <Text className="text-gray-500" variant="overline">
          {install?.id}
        </Text>
        <Heading>{install?.name}</Heading>
      </div>
      <div className="flex flex-col gap-1">
        <Text variant="caption">
          <b>App:</b> {install?.app?.name}
        </Text>

        <Text variant="caption">
          <b>Platform:</b> <InstallPlatform {...install} />
        </Text>

        <Text variant="caption">
          <b>Region:</b> <InstallRegion {...install} />
        </Text>

        <Text variant="caption">
          <b>Created by:</b> {install?.created_by?.email}
        </Text>
      </div>
    </div>
  )
}

export interface IInstallStatus {
  isCompact?: boolean
  isCompositeStatus?: boolean
  isStatusTextHidden?: boolean
  install: TInstall
}

export const InstallStatus: FC<IInstallStatus> = ({
  isCompact = false,
  isCompositeStatus = false,
  isStatusTextHidden = false,
  install,
}) => {
  const [status, setStatus] = useState(getFullInstallStatus(install))
  const fetchStatus = () => {
    fetch(`/api/${install?.org_id}/${install?.id}/status`)
      .then((res) => res.json().then((s) => setStatus(s)))
      .catch(console.error)
  }

  useEffect(() => {
    fetchStatus()
  }, [])

  let pollStatus: NodeJS.Timeout
  useEffect(() => {
    pollStatus = setInterval(fetchStatus, 5000)
    return () => clearInterval(pollStatus)
  }, [status])

  return isCompositeStatus ? (
    <Status
      status={status?.installStatus?.status}
      description={!isCompact && status?.installStatus?.status_description}
      isStatusTextHidden={isStatusTextHidden}
    />
  ) : (
    <div className="flex flex-auto gap-6">
      <Status
        status={status?.sandboxStatus?.status}
        description={!isCompact && status?.sandboxStatus?.status_description}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Sandbox'}
      />
      <Status
        status={status?.componentStatus?.status}
        description={!isCompact && status?.componentStatus?.status_description}
        isLabelStatusText={isCompact}
        isStatusTextHidden={isStatusTextHidden}
        label={!isStatusTextHidden && 'Components'}
      />
    </div>
  )
}

export const SandboxDetails: FC<TSandboxConfig> = ({
  artifacts,
  connected_github_vcs_config,
  public_git_vcs_config,
  terraform_version,
  variables,
}) => {
  const isGithubConnected = Boolean(connected_github_vcs_config)
  const repo = connected_github_vcs_config || public_git_vcs_config

  return (
    <div className="flex flex-col gap-2">
      <Heading>Sandbox</Heading>
      <span>
        <Text variant="caption" className="flex items-center gap-1">
          <b>Repo:</b>{' '}
          <Link
            href={`https://github.com/${repo?.repo}`}
            target="_blank"
            rel="noreferrer"
          >
            {isGithubConnected ? <FaGithub /> : <FaGitAlt />}
            {repo?.repo}
          </Link>
        </Text>

        <Text variant="caption" className="flex items-center gap-1">
          <b>Directory:</b> {repo?.directory}
        </Text>

        <Text variant="caption" className="flex items-center gap-1">
          <b>Branch:</b> {repo?.branch}
        </Text>

        <Text variant="caption" className="flex items-center gap-1">
          <b>Terraform:</b> {terraform_version}
        </Text>
      </span>

      <Heading className="flex items-center gap-2" variant="subheading">
        Variables{' '}
        <Link
          className="text-sm"
          href="https://docs.nuon.co/guides/sandboxes#sandbox-outputs"
          target="_blank"
          rel="noreferrer"
        >
          <GoInfo />
        </Link>
      </Heading>

      <Code variant="preformated">{JSON.stringify(variables, null, 2)}</Code>
    </div>
  )
}

export const AwsAccount: FC<TInstallAwsAccount> = ({
  iam_role_arn,
  region,
}) => {
  return (
    <>
      <Text variant="caption" className="flex items-center gap-1">
        <b>Region:</b> <InstallRegion {...{ aws_account: { region } }} />
      </Text>
      <Text variant="caption" className="flex items-center gap-1">
        <b>IAM role ARN:</b> {iam_role_arn}
      </Text>
    </>
  )
}

export const AzureAccount: FC<TInstallAzureAccount> = ({
  location,
  subscription_id,
}) => {
  return (
    <>
      <Text variant="caption" className="flex items-center gap-1">
        <b>Location:</b> <InstallRegion {...{ azure_account: { location } }} />
      </Text>
      <Text variant="caption" className="flex items-center gap-1">
        <b>Subscription ID:</b> {subscription_id}
      </Text>
    </>
  )
}

export const CloudDetails: FC<TInstall> = ({
  app_runner_config,
  app_sandbox_config,
  aws_account,
  azure_account,
}) => {
  return (
    <div className="flex flex-col gap-2">
      <Heading>Cloud account</Heading>

      <span>
        <Text variant="caption" className="flex items-center gap-1">
          <b>Platform:</b> <InstallPlatform {...app_sandbox_config} />
        </Text>

        <Text variant="caption" className="flex items-center gap-1">
          <b>Runner type:</b> {app_runner_config?.app_runner_type}
        </Text>

        {app_runner_config?.cloud_platform === 'azure' ? (
          <AzureAccount {...azure_account} />
        ) : (
          <AwsAccount {...aws_account} />
        )}
      </span>

      {app_runner_config?.cloud_platform !== 'azure' ? (
        <Policies {...app_sandbox_config?.artifacts} />
      ) : null}
    </div>
  )
}

export const Policies: FC<{
  deprovision_policy?: string
  provision_policy?: string
  trust_policy?: string
}> = ({
  deprovision_policy = '',
  provision_policy = '',
  trust_policy = '',
}) => {
  return (
    <>
      <Heading className="flex items-center gap-2" variant="subheading">
        Policies{' '}
        <Link
          className="text-sm"
          href="https://docs.nuon.co/guides/install-access-permissions"
          target="_blank"
          rel="noreferrer"
        >
          <GoInfo />
        </Link>
      </Heading>
      <div className="flex items-center gap-4">
        <Link
          className="text-sm"
          href={trust_policy}
          target="_blank"
          rel="noreferrer"
        >
          Trust policy
        </Link>

        <Link
          className="text-sm"
          href={provision_policy}
          target="_blank"
          rel="noreferrer"
        >
          Provision policy
        </Link>

        <Link
          className="text-sm"
          href={deprovision_policy}
          target="_blank"
          rel="noreferrer"
        >
          Deprovision policy
        </Link>
      </div>
    </>
  )
}

export const InstallTitle: FC<TInstall> = ({ name, id }) => {
  return <></>
}

export const InstallPlatform: FC<
  TSandboxConfig & { hasTextHidden?: boolean }
> = ({ cloud_platform: platform, hasTextHidden = false }) => {
  return (
    <span className="flex items-center gap-2">
      {platform === 'azure' ? (
        <>
          <VscAzure className="text-md" /> {!hasTextHidden && 'Azure'}
        </>
      ) : (
        <>
          <FaAws className="text-xl mb-[-4px]" /> {!hasTextHidden && 'Amazon'}
        </>
      )}
    </span>
  )
}

export const InstallRegion: FC<TInstall> = ({ aws_account, azure_account }) => {
  const region = azure_account
    ? AZURE_REGIONS.find((r) => r.value === azure_account?.location)
    : AWS_REGIONS.find((r) => r.value === aws_account?.region)

  return (
    <span className="flex gap-2">
      {getFlagEmoji(region?.iconVariant?.substring(5))} {region?.text}
    </span>
  )
}

export const InstallHeading: FC<TInstall> = ({
  app,
  app_sandbox_config,
  id,
  name,
  ...install
}) => {
  console.log('install', name)

  return (
    <div className="flex flex-wrap gap-8 items-end border-b pb-6">
      <div className="flex flex-col flex-auto gap-2">
        <span className="flex flex-col gap-0">
          <Text variant="overline">{id}</Text>
          <Heading level={1} variant="title">
            {name}
          </Heading>
        </span>

        <Text className="flex flex-wrap gap-4 items-center" variant="caption">
          <Text variant="status">{app?.name}</Text>{' '}
          <InstallPlatform {...app_sandbox_config} hasTextHidden />{' '}
          <InstallRegion {...install} />
        </Text>
      </div>

      <InstallStatus install={{ id, ...install }} />
    </div>
  )
}
