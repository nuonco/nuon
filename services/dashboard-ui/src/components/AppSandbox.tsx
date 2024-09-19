import React, { type FC } from 'react'
import { FaGitAlt, FaGithub } from 'react-icons/fa'
import { Config, ConfigContent, Text, Link } from '@/components'
import type { TAppSandboxConfig } from '@/types'

export interface IAppSandboxConfig {
  sandboxConfig: TAppSandboxConfig
}

export const AppSandboxConfig: FC<IAppSandboxConfig> = ({ sandboxConfig }) => {
  const isGithubConnected = Boolean(sandboxConfig.connected_github_vcs_config)
  const repo =
    sandboxConfig.connected_github_vcs_config ||
    sandboxConfig.public_git_vcs_config

  return (
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
            {repo?.repo}
          </Link>
        }
      />
      <ConfigContent label="Directory" value={repo?.directory} />
      <ConfigContent label="Branch" value={repo?.branch} />
      <ConfigContent
        label="Terraform"
        value={sandboxConfig.terraform_version}
      />
    </Config>
  )
}

export interface IAppSandboxVariables {
  variables: TAppSandboxConfig['variables']
}

export const AppSandboxVariables: FC<IAppSandboxVariables> = ({
  variables,
}) => {
  const variableKeys = Object.keys(variables)
  const isEmpty = variableKeys.length === 0

  return isEmpty ? null : (
    <div className="flex flex-col gap-4">
      <div className="">
        <Text className="text-sm !font-medium leading-normal">Variables</Text>
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
            <Text className="font-mono text-sm">{key}</Text>
            <Text className="text-sm font-mono break-all col-span-2">
              {variables[key]}
            </Text>
          </div>
        ))}
      </div>
    </div>
  )
}
