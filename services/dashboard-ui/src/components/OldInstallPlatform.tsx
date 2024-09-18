'use client'

import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { GoInfo } from 'react-icons/go'
import { VscAzure } from 'react-icons/vsc'
import { Card, Code, Heading, Link, Text } from '@/components'
import { useInstallContext } from '@/context'
import { AWS_REGIONS, AZURE_REGIONS, getFlagEmoji } from '@/utils'

export const InstallConfigAWS: FC = () => {
  const {
    install: {
      aws_account: { iam_role_arn },
    },
  } = useInstallContext()

  return (
    <>
      <Text variant="caption" className="flex flex-nowrap items-center gap-1">
        <b className="flex-[1_0_fit-content]">IAM role ARN:</b>{' '}
        <Code variant="inline" className="shrink">
          {iam_role_arn}
        </Code>
      </Text>
    </>
  )
}

export const InstallConfigAzure: FC = () => {
  const {
    install: {
      azure_account: { subscription_id },
    },
  } = useInstallContext()

  return (
    <>
      <Text variant="caption" className="flex items-center gap-1">
        <b>Subscription ID:</b> {subscription_id}
      </Text>
    </>
  )
}

export const InstallConfig: FC = () => {
  const { install } = useInstallContext()
  const isAzure = Boolean(install.azure_account)

  return (
    <>
      <Text variant="caption" className="flex items-center gap-1">
        <b>{isAzure ? 'Location' : 'Region'}:</b> <InstallRegion />
      </Text>
      {isAzure ? <InstallConfigAzure /> : <InstallConfigAWS />}
    </>
  )
}

export const Policies: FC = () => {
  const {
    install: {
      app_sandbox_config: {
        artifacts: { deprovision_policy, provision_policy, trust_policy },
      },
    },
  } = useInstallContext()

  return (
    <Text variant="caption" className="flex items-center gap-4">
      <Link href={trust_policy} target="_blank" rel="noreferrer">
        Trust policy
      </Link>

      <Link href={provision_policy} target="_blank" rel="noreferrer">
        Provision policy
      </Link>

      <Link href={deprovision_policy} target="_blank" rel="noreferrer">
        Deprovision policy
      </Link>
    </Text>
  )
}

export const InstallPlatformType: FC<{
  isIconOnly?: boolean
}> = ({ isIconOnly = false }) => {
  const {
    install: {
      app_sandbox_config: { cloud_platform },
    },
  } = useInstallContext()

  return (
    <span className="flex items-center gap-2">
      {cloud_platform === 'azure' ? (
        <>
          <VscAzure className="text-md" /> {!isIconOnly && 'Azure'}
        </>
      ) : (
        <>
          <FaAws className="text-xl mb-[-4px]" /> {!isIconOnly && 'Amazon'}
        </>
      )}
    </span>
  )
}

export const InstallRegion: FC = () => {
  const {
    install: { aws_account, azure_account },
  } = useInstallContext()
  const region = azure_account
    ? AZURE_REGIONS.find((r) => r.value === azure_account?.location)
    : AWS_REGIONS.find((r) => r.value === aws_account?.region)

  return (
    <span className="flex gap-2">
      {getFlagEmoji(region?.iconVariant?.substring(5))} {region?.text}
    </span>
  )
}

//
// =========================================================================
// new stuff

export const InstallCloudPlatformDetailsCard: FC = () => {
  const { install } = useInstallContext()
  const isAzure = Boolean(install?.azure_account)
  return (
    <Card>
      <Heading>Cloud platform</Heading>

      <InstallCloudPlatformDetails />

      {isAzure ? null : (
        <>
          <Heading className="flex items-center gap-2" variant="subheading">
            Policies{' '}
            <Link
              className="text-sm"
              href="https://docs.nuon.co/guides/install-access-permissions"
              target="_blank"
              title="Install access permission documentation"
              rel="noreferrer"
            >
              <GoInfo />
            </Link>
          </Heading>
          <Policies />
        </>
      )}
    </Card>
  )
}

export const InstallCloudPlatformDetails: FC = () => {
  const {
    install: { app_runner_config },
  } = useInstallContext()

  return (
    <span>
      <Text variant="caption" className="flex items-center gap-1">
        <b>Platform:</b> <InstallPlatformType />
      </Text>

      <Text variant="caption" className="flex items-center gap-1">
        <b>Runner type:</b> {app_runner_config?.app_runner_type}
      </Text>

      <InstallConfig />
    </span>
  )
}
