import React, { type FC } from 'react'
import { FaGitAlt, FaGithub } from 'react-icons/fa'
import { Config, ConfigContent } from '@/components/old/Config'
import { ConfigurationVariables } from '@/components/old/ComponentConfig'
import { Link } from '@/components/old/Link'
import { ToolTip } from '@/components/old/ToolTip'
import { Text, Truncate } from '@/components/old/Typography'
import type { TAppSandboxConfig } from '@/types'

export interface IAppSandboxConfig {
  sandboxConfig: TAppSandboxConfig
}

export const AppSandboxConfig: FC<IAppSandboxConfig> = ({ sandboxConfig }) => {
  const isGithubConnected = Boolean(sandboxConfig?.connected_github_vcs_config)
  const repo =
    sandboxConfig?.connected_github_vcs_config ||
    sandboxConfig?.public_git_vcs_config

  return sandboxConfig ? (
    <Config>
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
                <Truncate variant="small">{repo?.repo}</Truncate>
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
      <ConfigContent
        label="Terraform"
        value={sandboxConfig?.terraform_version}
      />
    </Config>
  ) : (
    <Text>Missing app sandbox configuration</Text>
  )
}

export interface IAppSandboxVariables {
  heading?: string
  isNotTruncated?: boolean
  variables: TAppSandboxConfig['variables']
}

export const AppSandboxVariables: FC<IAppSandboxVariables> = ({
  heading = 'Variables',
  isNotTruncated = false,
  variables,
}) => {
  return (
    <ConfigurationVariables
      heading={heading}
      variables={variables}
      isNotTruncated={isNotTruncated}
    />
  )
  /* const variableKeys = Object.keys(variables || {})
   * const isEmpty = variableKeys.length === 0

   * return isEmpty ? null : (
   *   <div className="flex flex-col gap-4">
   *     <div className="">
   *       <Text className="text-sm !font-medium leading-normal">Variables</Text>
   *     </div>

   *     <div className="divide-y">
   *       <div className="grid grid-cols-3 gap-4 pb-3">
   *         <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
   *           Name
   *         </Text>
   *         <Text className="text-sm !font-medium text-cool-grey-600 dark:text-cool-grey-500">
   *           Value
   *         </Text>
   *       </div>

   *       {variableKeys.map((key, i) => (
   *         <div key={`${key}-${i}`} className="grid grid-cols-3 gap-4 py-3">
   *           <Text className="font-mono text-sm break-all">
   *             {key.length >= 15 ? (
   *               <ToolTip tipContent={key} alignment="right">
   *                 <Truncate variant="small">{key}</Truncate>
   *               </ToolTip>
   *             ) : (
   *               key
   *             )}
   *           </Text>
   *           <Text className="text-sm font-mono break-all col-span-2">
   *             {variables[key].length >= 20 ? (
   *               <ToolTip tipContent={variables[key]} alignment="right">
   *                 <Truncate variant="large">{variables[key]}</Truncate>
   *               </ToolTip>
   *             ) : (
   *               variables[key]
   *             )}
   *           </Text>
   *         </div>
   *       ))}
   *     </div>
   *   </div>
   * ) */
}

export const AppSandboxRepoDirLink: FC<{
  repoDirPath: string
  isGithubConnected: boolean
}> = ({ repoDirPath, isGithubConnected }) => {
  const urlParts = repoDirPath.split('/')

  return (
    <Link
      href={`https://github.com/${urlParts[0]}/${urlParts[1]}/tree/main/${urlParts[2]}`}
      target="_blank"
      rel="noreferrer"
    >
      {isGithubConnected ? <FaGithub /> : <FaGitAlt />}
      {repoDirPath.length >= 26 ? (
        <ToolTip alignment="right" tipContent={repoDirPath}>
          <Truncate>{repoDirPath}</Truncate>
        </ToolTip>
      ) : (
        <Text className="text-sm">{repoDirPath}</Text>
      )}
    </Link>
  )
}
