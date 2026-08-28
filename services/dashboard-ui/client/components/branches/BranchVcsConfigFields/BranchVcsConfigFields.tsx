import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { ConnectGithubButton } from '@/components/vcs-connections/ConnectGithub'
import type { TVCSConnection, TVCSConnectionRepo, TVCSBranch } from '@/types'

export interface IBranchVcsConfigFields {
  vcsConnections: TVCSConnection[]
  repos: TVCSConnectionRepo[]
  branches: TVCSBranch[]
  loadingRepos: boolean
  loadingBranches: boolean
  reposError: string | null
  branchesError: string | null
  selectedVcsConnectionId: string
  onVcsConnectionChange: (id: string) => void
  selectedRepo: TVCSConnectionRepo | null
  onRepoChange: (repo: TVCSConnectionRepo | null) => void
  selectedBranch: string
  onBranchChange: (branch: string) => void
  directory: string
  onDirectoryChange: (directory: string) => void
  // Omit both to hide the path filter field, for consumers whose config has no
  // path filter (e.g. the app installs config).
  pathFilter?: string
  onPathFilterChange?: (pathFilter: string) => void
  isSubmitting: boolean
}

const TEXT_BUTTON_CLASS =
  'w-fit !p-0 !h-auto !bg-transparent !border-none font-medium text-primary-600 dark:text-primary-500 hover:text-primary-800 hover:dark:text-primary-400'

const publicRepoEntry = (value: string): TVCSConnectionRepo | null => {
  const trimmed = value.trim()
  if (!trimmed) return null
  const fullName = trimmed
    .replace(/^https?:\/\/github\.com\//, '')
    .replace(/\.git$/, '')
    .replace(/\/$/, '')
  return {
    full_name: fullName,
    name: fullName.split('/')[1] || fullName,
    private: false,
  } as TVCSConnectionRepo
}

const FieldSkeleton = ({
  htmlFor,
  label,
}: {
  htmlFor: string
  label: string
}) => (
  <div className="flex flex-col gap-1">
    <Label htmlFor={htmlFor}>
      <Text variant="body" className="font-medium">
        {label}
      </Text>
    </Label>
    <Skeleton height="36px" />
  </div>
)

export const BranchVcsConfigFields = ({
  vcsConnections,
  repos,
  branches,
  loadingRepos,
  loadingBranches,
  reposError,
  branchesError,
  selectedVcsConnectionId,
  onVcsConnectionChange,
  selectedRepo,
  onRepoChange,
  selectedBranch,
  onBranchChange,
  directory,
  onDirectoryChange,
  pathFilter,
  onPathFilterChange,
  isSubmitting,
}: IBranchVcsConfigFields) => {
  const hasConnections = vcsConnections.length > 0
  const [usePublicRepo, setUsePublicRepo] = useState(!hasConnections)
  const [publicRepoText, setPublicRepoText] = useState(
    selectedRepo?.full_name ?? ''
  )

  useEffect(() => {
    if (!usePublicRepo || publicRepoText) return
    if (selectedRepo && !selectedRepo.private) {
      setPublicRepoText(selectedRepo.full_name)
    }
  }, [usePublicRepo, selectedRepo?.full_name])

  const switchMode = (toPublic: boolean) => {
    setUsePublicRepo(toPublic)
    if (toPublic) {
      const keep = selectedRepo && !selectedRepo.private ? selectedRepo : null
      setPublicRepoText(keep?.full_name ?? '')
      onRepoChange(keep)
      return
    }
    setPublicRepoText('')
    onRepoChange(repos[0] || null)
  }

  if (usePublicRepo) {
    return (
      <>
        {!hasConnections && (
          <Banner theme="warn">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <Text>
                No GitHub connections yet. Public repositories work without one.
              </Text>
              <ConnectGithubButton
                type="button"
                variant="secondary"
                size="sm"
                disabled={isSubmitting}
              >
                Connect GitHub
              </ConnectGithubButton>
            </div>
          </Banner>
        )}

        <Input
          id="public-repo"
          type="text"
          value={publicRepoText}
          onChange={(e) => {
            setPublicRepoText(e.target.value)
            onRepoChange(publicRepoEntry(e.target.value))
          }}
          placeholder="owner/repo"
          required
          disabled={isSubmitting}
          labelProps={{ labelText: 'Public repository' }}
          helperText="Nuon clones this anonymously, so the repository must be public."
        />

        <Input
          id="git-branch"
          type="text"
          value={selectedBranch}
          onChange={(e) => onBranchChange(e.target.value)}
          placeholder="main"
          required
          disabled={isSubmitting}
          labelProps={{ labelText: 'Git branch' }}
        />

        <Input
          id="directory"
          type="text"
          value={directory}
          onChange={(e) => onDirectoryChange(e.target.value)}
          placeholder="."
          required
          disabled={isSubmitting}
          labelProps={{ labelText: 'Directory' }}
          helperText='Path to your application config (use "." for root)'
        />

        {onPathFilterChange && (
          <Input
            id="path-filter"
            type="text"
            value={pathFilter ?? ''}
            onChange={(e) => onPathFilterChange(e.target.value)}
            placeholder="^(src/|config/).*"
            disabled={isSubmitting}
            labelProps={{ labelText: 'Path filter (optional)' }}
            helperText="Regex pattern to filter which file changes trigger workflow runs"
          />
        )}

        <Banner theme="info">
          Pushes and pull requests only trigger runs if this repository&apos;s
          GitHub organization has a Nuon connection. Without one, trigger runs
          manually and Nuon cannot post commit statuses or PR comments.
        </Banner>

        {hasConnections && (
          <Button
            type="button"
            variant="ghost"
            size="sm"
            className={TEXT_BUTTON_CLASS}
            disabled={isSubmitting}
            onClick={() => switchMode(false)}
          >
            Use a connected repository instead
          </Button>
        )}
      </>
    )
  }

  return (
    <>
      {vcsConnections.length > 1 && (
        <Select
          id="vcs-connection"
          value={selectedVcsConnectionId}
          onChange={(value) => onVcsConnectionChange(value)}
          disabled={isSubmitting || loadingRepos}
          options={vcsConnections.map((conn) => ({
            value: conn.id,
            label:
              conn.github_account_name || conn.github_install_id || conn.id,
          }))}
          labelProps={{ labelText: 'VCS connection' }}
        />
      )}

      {reposError && <Banner theme="error">{reposError}</Banner>}

      {loadingRepos ? (
        <FieldSkeleton htmlFor="repo" label="Repository" />
      ) : reposError ? (
        <Banner theme="error">Failed to load repositories</Banner>
      ) : repos.length === 0 ? (
        <Banner theme="warn">
          No connected repositories found. Update your GitHub connection to
          grant access to repositories.
        </Banner>
      ) : (
        <Select
          id="repo"
          value={selectedRepo?.full_name || ''}
          onChange={(value) => {
            const repo = repos.find((r) => r.full_name === value)
            onRepoChange(repo || null)
          }}
          required
          disabled={isSubmitting || loadingRepos || loadingBranches}
          options={repos.map((repo) => ({
            value: repo.full_name,
            label: repo.name,
            badge: repo.private
              ? { label: 'private' }
              : { label: repo.full_name.split('/')[0] },
          }))}
          labelProps={{ labelText: 'Repository' }}
          searchable
        />
      )}

      {!loadingRepos && branchesError && (
        <Banner theme="error">{branchesError}</Banner>
      )}

      {!loadingRepos &&
        (loadingBranches ? (
          <FieldSkeleton htmlFor="git-branch" label="Git branch" />
        ) : branchesError ? (
          <Input
            id="git-branch"
            type="text"
            value={selectedBranch}
            onChange={(e) => onBranchChange(e.target.value)}
            placeholder="main"
            required
            disabled={isSubmitting}
            labelProps={{ labelText: 'Git branch' }}
          />
        ) : branches.length === 0 && selectedRepo ? (
          <Input
            id="git-branch"
            type="text"
            value={selectedBranch}
            onChange={(e) => onBranchChange(e.target.value)}
            placeholder="main"
            required
            disabled={isSubmitting}
            labelProps={{ labelText: 'Git branch' }}
            helperText="No branches found. Enter branch name manually."
          />
        ) : branches.length > 0 ? (
          <Select
            id="git-branch"
            value={selectedBranch}
            onChange={(value) => onBranchChange(value)}
            required
            disabled={isSubmitting || loadingBranches}
            options={branches.map((b) => ({
              value: b.name,
              label: b.name,
            }))}
            labelProps={{ labelText: 'Git branch' }}
            searchable
          />
        ) : (
          <Input
            id="git-branch"
            type="text"
            value={selectedBranch}
            onChange={(e) => onBranchChange(e.target.value)}
            placeholder="main"
            required
            disabled={isSubmitting}
            labelProps={{ labelText: 'Git branch' }}
          />
        ))}

      <Input
        id="directory"
        type="text"
        value={directory}
        onChange={(e) => onDirectoryChange(e.target.value)}
        placeholder="."
        required
        disabled={isSubmitting}
        labelProps={{ labelText: 'Directory' }}
        helperText='Path to your application config (use "." for root)'
      />

      {onPathFilterChange && (
        <Input
          id="path-filter"
          type="text"
          value={pathFilter ?? ''}
          onChange={(e) => onPathFilterChange(e.target.value)}
          placeholder="^(src/|config/).*"
          disabled={isSubmitting}
          labelProps={{ labelText: 'Path filter (optional)' }}
          helperText="Regex pattern to filter which file changes trigger workflow runs"
        />
      )}

      <Button
        type="button"
        variant="ghost"
        size="sm"
        className={TEXT_BUTTON_CLASS}
        disabled={isSubmitting}
        onClick={() => switchMode(true)}
      >
        Use a public repository instead
      </Button>
    </>
  )
}
