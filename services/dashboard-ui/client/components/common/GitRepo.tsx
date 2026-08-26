import type { TVCSGit, TVCSGitHub } from '@/types/ctl-api.types'
import { LabeledValue } from './LabeledValue'
import { Link } from './Link'
import { Text } from './Text'

export interface IGitRepo {
  vcsConfig: TVCSGitHub | TVCSGit
  isAutoGrid?: boolean
}

const buildDirectoryUrl = (
  repo: string,
  branch: string | undefined,
  directory: string
): string | null => {
  if (!branch) return null

  const cleanRepo = repo.replace(/\.git$/, '')

  if (cleanRepo.includes('github.com')) {
    return `${cleanRepo}/tree/${branch}/${directory}`
  } else if (cleanRepo.includes('gitlab.com')) {
    return `${cleanRepo}/-/tree/${branch}/${directory}`
  } else if (cleanRepo.includes('bitbucket.org')) {
    return `${cleanRepo}/src/${branch}/${directory}`
  } else {
    return `https://github.com/${cleanRepo}/tree/${branch}/${directory}`
  }

  return `${cleanRepo}/tree/${branch}/${directory}`
}

export const GitRepo = ({ vcsConfig, isAutoGrid = false }: IGitRepo) => {
  const isGitHub = 'repo_owner' in vcsConfig

  const directoryUrl =
    vcsConfig?.repo && vcsConfig?.directory
      ? buildDirectoryUrl(vcsConfig.repo, vcsConfig.branch, vcsConfig.directory)
      : null

  let columnCount = 0
  if (vcsConfig?.repo) columnCount++
  if (isGitHub && vcsConfig.repo_owner) columnCount++
  if (isGitHub && vcsConfig.repo_name) columnCount++
  if (vcsConfig?.branch) columnCount++
  if (vcsConfig?.directory) columnCount++

  const gridTemplateColumns = isAutoGrid
    ? 'auto auto auto'
    : columnCount > 0
      ? `2fr ${Array(columnCount - 1)
          .fill('1fr')
          .join(' ')}`
      : undefined

  return (
    <div className="grid gap-6" style={{ gridTemplateColumns }}>
      {vcsConfig?.repo && (
        <LabeledValue label="Repository">
          <Link href={`https://github.com/${vcsConfig.repo}`} isExternal>
            {vcsConfig.repo}
          </Link>
        </LabeledValue>
      )}

      {isGitHub && vcsConfig.repo_owner && (
        <LabeledValue label="Owner">
          <Text variant="subtext">{vcsConfig.repo_owner}</Text>
        </LabeledValue>
      )}

      {isGitHub && vcsConfig.repo_name && (
        <LabeledValue label="Repository Name">
          <Text variant="subtext">{vcsConfig.repo_name}</Text>
        </LabeledValue>
      )}

      {vcsConfig?.branch && (
        <LabeledValue label="Branch">
          <Text variant="subtext">{vcsConfig.branch}</Text>
        </LabeledValue>
      )}

      {vcsConfig?.directory && (
        <LabeledValue label="Directory">
          <Text variant="subtext">
            {directoryUrl ? (
              <Link href={directoryUrl} isExternal variant="inline">
                {vcsConfig.directory}
              </Link>
            ) : (
              vcsConfig.directory
            )}
          </Text>
        </LabeledValue>
      )}
    </div>
  )
}
