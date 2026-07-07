import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Label } from '@/components/common/form/Label'
import { Select } from '@/components/common/form/Select'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
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
  pathFilter: string
  onPathFilterChange: (pathFilter: string) => void
  isSubmitting: boolean
}

const FieldSkeleton = ({ htmlFor, label }: { htmlFor: string; label: string }) => (
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
  if (vcsConnections.length === 0) {
    return (
      <Banner theme="warn">
        No VCS connections found. Connect your GitHub account first.
      </Banner>
    )
  }

  return (
    <>
      {vcsConnections.length > 1 && (
        <Select
          id="vcs-connection"
          value={selectedVcsConnectionId}
          onChange={(e) => onVcsConnectionChange(e.target.value)}
          disabled={isSubmitting || loadingRepos}
          options={vcsConnections.map((conn) => ({
            value: conn.id,
            label: conn.github_account_name || conn.github_install_id || conn.id,
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
          No connected repositories found. Update your GitHub connection to grant
          access to repositories.
        </Banner>
      ) : (
        <Select
          id="repo"
          value={selectedRepo?.full_name || ''}
          onChange={(e) => {
            const repo = repos.find((r) => r.full_name === e.target.value)
            onRepoChange(repo || null)
          }}
          required
          disabled={isSubmitting || loadingRepos || loadingBranches}
          options={repos.map((repo) => ({
            value: repo.full_name,
            label: repo.full_name,
            badge: repo.private ? { label: 'private' } : undefined,
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
            onChange={(e) => onBranchChange(e.target.value)}
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

      <Input
        id="path-filter"
        type="text"
        value={pathFilter}
        onChange={(e) => onPathFilterChange(e.target.value)}
        placeholder="^(src/|config/).*"
        disabled={isSubmitting}
        labelProps={{ labelText: 'Path filter (optional)' }}
        helperText="Regex pattern to filter which file changes trigger workflow runs"
      />
    </>
  )
}
