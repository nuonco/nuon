'use client'

import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import { QuestionMark } from '@phosphor-icons/react'
import { Link, Text, ToolTip } from '@/components'
import type { TInstall, TSandboxConfig } from '@/types'
import { AWS_REGIONS, AZURE_REGIONS, getFlagEmoji } from '@/utils'

const InstallPlatform: FC<{ platform: 'aws' | 'azure' | string }> = ({
  platform,
}) => {
  return platform === 'azure' ? (
    <>
      <VscAzure className="text-md" /> {'Azure'}
    </>
  ) : (
    <>
      <FaAws className="text-xl mb-[-4px]" /> {'Amazon'}
    </>
  )
}

const AzureAccount: FC<Pick<TInstall, 'azure_account'>> = ({
  azure_account,
}) => {
  const region = AZURE_REGIONS.find((r) => r.value === azure_account.location)
  return (
    <>
      <span className="flex flex-col gap-2">
        <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          Location
        </Text>
        <Text className="text-sm">
          {getFlagEmoji(region.iconVariant?.substring(5))} {region.text}
        </Text>
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          Subscription ID
        </Text>
        <ToolTip tipContent={azure_account.subscription_id} alignment="right">
          <Text className="truncate text-ellipsis w-16 text-sm font-mono">
            {azure_account.subscription_id}
          </Text>
        </ToolTip>
      </span>
    </>
  )
}

const AWSAccount: FC<Pick<TInstall, 'aws_account'>> = ({ aws_account }) => {
  const region = AWS_REGIONS.find((r) => r.value === aws_account.region)

  return (
    <>
      <span className="flex flex-col gap-2">
        <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          Region
        </Text>
        <Text className="text-sm">
          {getFlagEmoji(region.iconVariant?.substring(5))} {region.text}
        </Text>
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          IAM role ARN
        </Text>
        <ToolTip tipContent={aws_account.iam_role_arn} alignment="right">
          <Text className="truncate text-ellipsis w-16 text-sm font-mono">
            {aws_account.iam_role_arn}
          </Text>
        </ToolTip>
      </span>
    </>
  )
}

const AWSPolicies: FC<Pick<TSandboxConfig, 'artifacts'>> = ({ artifacts }) => {
  return (
    <div className="flex flex-col gap-4">
      <Text className="flex items-center gap-2 text-sm !font-medium leading-normal">
        Policies{' '}
        <Link
          className="text-sm"
          href="https://docs.nuon.co/guides/install-access-permissions"
          target="_blank"
          title="Install access permission documentation"
          rel="noreferrer"
        >
          <QuestionMark />
        </Link>
      </Text>
      <Text className="flex items-center gap-4 text-sm">
        <Link href={artifacts?.trust_policy} target="_blank" rel="noreferrer">
          Trust policy
        </Link>

        <Link
          href={artifacts?.provision_policy}
          target="_blank"
          rel="noreferrer"
        >
          Provision policy
        </Link>

        <Link
          href={artifacts?.deprovision_policy}
          target="_blank"
          rel="noreferrer"
        >
          Deprovision policy
        </Link>
      </Text>
    </div>
  )
}

export interface IInstallCloudPlatform {
  install: TInstall
}

export const InstallCloudPlatform: FC<IInstallCloudPlatform> = ({
  install,
}) => {
  const {
    aws_account,
    azure_account,
    app_runner_config: { cloud_platform },
  } = install
  const isAWS = Boolean(aws_account)

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col md:flex-row gap-4">
        <span className="flex flex-col gap-2">
          <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
            Platform
          </Text>
          <Text className="text-sm">
            <InstallPlatform platform={cloud_platform} />
          </Text>
        </span>

        <span className="flex flex-col gap-2">
          <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
            Runner type
          </Text>
          <Text className="text-sm">
            {install.app_runner_config?.app_runner_type}
          </Text>
        </span>

        {isAWS ? <AWSAccount {...install} /> : <AzureAccount {...install} />}
      </div>

      {isAWS && (
        <AWSPolicies artifacts={install?.app_sandbox_config?.artifacts} />
      )}
    </div>
  )
}
