'use client'

import React, { type FC } from 'react'
import { FaGitAlt, FaGithub } from 'react-icons/fa'
import { GoInfo } from 'react-icons/go'
import { Card, Code, Heading, Link, Status, Text } from '@/components'
import { useInstallContext, useSandboxRunContext } from '@/context'

export const InstallSandboxDetailsCard: FC = () => {
  return (
    <Card>
      <Heading>Sandbox</Heading>
      <InstallSandboxDetails />

      <Heading className="flex items-center gap-2" variant="subheading">
        Variables{' '}
        <Link
          className="text-sm"
          href="https://docs.nuon.co/guides/sandboxes#sandbox-outputs"
          target="_blank"
          title="Sandbox outputs documentation"
          rel="noreferrer"
        >
          <GoInfo />
        </Link>
      </Heading>

      <InstallSandboxVariables />
    </Card>
  )
}

export const InstallSandboxDetails: FC = () => {
  const {
    install: {
      app_sandbox_config: {
        connected_github_vcs_config,
        public_git_vcs_config,
        terraform_version,
      },
    },
  } = useInstallContext()
  const isGithubConnected = Boolean(connected_github_vcs_config)
  const repo = connected_github_vcs_config || public_git_vcs_config

  return (
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
  )
}

export const InstallSandboxVariables: FC = () => {
  const {
    install: {
      app_sandbox_config: { variables },
    },
  } = useInstallContext()
  return <Code variant="preformated">{JSON.stringify(variables, null, 2)}</Code>
}

export interface IInstallSandboxRunStatus {
  isCompact?: boolean
  isStatusTextHidden?: boolean
  showDescription?: boolean
}

export const InstallSandboxRunStatus: FC<IInstallSandboxRunStatus> = ({
  isCompact = false,
  isStatusTextHidden = false,
  showDescription = false,
}) => {
  const { run } = useSandboxRunContext()

  return (
    <Status
      status={run.status}
      description={showDescription && run.status_description}
      label={isCompact && run.run_type}
      isLabelStatusText={isCompact}
      isStatusTextHidden={isStatusTextHidden}
    />
  )
}
