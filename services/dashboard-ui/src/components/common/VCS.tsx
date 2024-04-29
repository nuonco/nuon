import React, { type FC } from 'react'
import { FaGitAlt, FaGithub } from 'react-icons/fa'
import { Link, Text } from '@/components'
import type { TVCSGit, TVCSGitHub } from '@/types'

export interface IVCS {
  connected_github_vcs_config?: TVCSGitHub
  public_git_vcs_config?: TVCSGit
}

export const VCS: FC<IVCS> = ({
  connected_github_vcs_config,
  public_git_vcs_config,
}) => {
  const isGithubConnected = Boolean(connected_github_vcs_config)
  const repo = connected_github_vcs_config || public_git_vcs_config

  return (
    <div className="flex flex-col">
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
    </div>
  )
}
