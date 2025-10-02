'use client'

import React, { type FC } from 'react'
import { FaAws } from 'react-icons/fa'
import { VscAzure } from 'react-icons/vsc'
import {  ConfigContent } from '@/components/Config'
import { ToolTip } from '@/components/ToolTip'
import {  Text } from '@/components/Typography'
import { AWS_REGIONS, AZURE_REGIONS } from "@/configs/cloud-regions"
import type { TInstall } from '@/types'
import { getFlagEmoji } from '@/utils'

export const InstallPlatform: FC<{ platform: 'aws' | 'azure' | string }> = ({
  platform,
}) => {
  return (
    <span className="flex gap-2 items-center">
      {platform === 'azure' ? (
        <>
          <VscAzure className="text-md" /> {'Azure'}
        </>
      ) : (
        <>
          <FaAws className="text-xl mb-[-4px]" /> {'Amazon'}
        </>
      )}
    </span>
  )
}

const AzureAccount: FC<Pick<TInstall, 'azure_account'>> = ({
  azure_account,
}) => {
  const region = AZURE_REGIONS.find((r) => r.value === azure_account.location)
  return (
    <>
      <ConfigContent
        label="Location"
        value={
          <span className="flex gap-2">
            {getFlagEmoji(region?.iconVariant?.substring(5))} {region?.text}
          </span>
        }
      />
      <ConfigContent
        label="Subscription ID"
        value={
          <ToolTip tipContent={azure_account.subscription_id} alignment="right">
            <Text className="truncate text-ellipsis w-16 text-sm font-mono">
              {azure_account.subscription_id}
            </Text>
          </ToolTip>
        }
      />
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

      {/* <span className="flex flex-col gap-2">
          <Text className="text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500">
          IAM role ARN
          </Text>

          <Code className="w-fit">
          <ClickToCopy
          className="!items-start"
          noticeClassName="-top-[5px] right-5"
          >
          {aws_account.iam_role_arn}
          </ClickToCopy>
          </Code>
          </span> */}
    </>
  )
}

/* const AWSPolicies: FC<Pick<TSandboxConfig, 'artifacts'>> = ({ artifacts }) => {
 *   return (
 *     <div className="flex flex-col gap-4">
 *       <Text className="flex items-center gap-2 text-sm !font-medium leading-normal">
 *         Policies{' '}
 *         <Link
 *           className="text-sm"
 *           href="https://docs.nuon.co/guides/install-access-permissions"
 *           target="_blank"
 *           title="Install access permission documentation"
 *           rel="noreferrer"
 *         >
 *           <QuestionMark />
 *         </Link>
 *       </Text>
 *       <Text className="flex items-center gap-4 text-sm">
 *         <Link href={artifacts?.trust_policy} target="_blank" rel="noreferrer">
 *           Trust policy
 *         </Link>
 * 
 *         <Link
 *           href={artifacts?.provision_policy}
 *           target="_blank"
 *           rel="noreferrer"
 *         >
 *           Provision policy
 *         </Link>
 * 
 *         <Link
 *           href={artifacts?.deprovision_policy}
 *           target="_blank"
 *           rel="noreferrer"
 *         >
 *           Deprovision policy
 *         </Link>
 *       </Text>
 *     </div>
 *   )
 * }
 * 
 * export interface IInstallCloudPlatform {
 *   install: TInstall
 * }
 * 
 * export const InstallCloudPlatform: FC<IInstallCloudPlatform> = ({
 *   install,
 * }) => {
 *   const {
 *     app_runner_config: { cloud_platform },
 *   } = install
 *   const isAWS = Boolean(install.aws_account)
 * 
 *   return (
 *     <div className="flex flex-col gap-6">
 *       <Config>
 *         <ConfigContent
 *           label="Platform"
 *           value={<InstallPlatform platform={cloud_platform} />}
 *         />
 *         <ConfigContent
 *           label="Runner type"
 *           value={install.app_runner_config?.app_runner_type}
 *         />
 *         {isAWS ? <AWSAccount {...install} /> : <AzureAccount {...install} />}
 *       </Config>
 * 
 *       {isAWS && (
 *         <AWSPolicies artifacts={install?.app_sandbox_config?.artifacts} />
 *       )}
 *     </div>
 *   )
 * } */
