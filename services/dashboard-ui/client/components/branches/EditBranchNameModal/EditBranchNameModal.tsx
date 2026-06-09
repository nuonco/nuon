import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TAppBranch,
  TAppBranchConfig,
  TVCSConnectionRepo,
  TVCSBranch,
  TVCSConnection,
} from '@/types'

export interface IEditBranchNameModalSubmitData {
  branchName: string
  useVcs: boolean
  selectedVcsConnectionId: string
  selectedRepo: TVCSConnectionRepo | null
  selectedBranch: string
  directory: string
  pathFilter: string
}

interface IEditBranchNameModal extends Omit<IModal, 'onSubmit'> {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
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
  isSubmitting: boolean
  validationError: string | null
  onSubmit: (data: IEditBranchNameModalSubmitData) => void
  onCancel: () => void
}

export const EditBranchNameModal = ({
  branch,
  currentConfig,
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
  isSubmitting,
  validationError: externalValidationError,
  onSubmit,
  onCancel,
  ...props
}: IEditBranchNameModal) => {
  const [branchName, setBranchName] = useState(branch.name || '')
  const [useVcs, setUseVcs] = useState(
    !!(currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config)
  )
  const [directory, setDirectory] = useState(
    currentConfig?.connected_github_vcs_config?.directory ||
      currentConfig?.public_git_vcs_config?.directory ||
      '.'
  )
  const [pathFilter, setPathFilter] = useState(
    currentConfig?.connected_github_vcs_config?.path_filter ||
      currentConfig?.public_git_vcs_config?.path_filter ||
      ''
  )
  const [validationError, setValidationError] = useState<string | null>(null)

  const displayError = externalValidationError || validationError

  const handleSubmit = () => {
    setValidationError(null)

    if (!branchName.trim()) {
      setValidationError('Branch name cannot be empty')
      return
    }

    if (useVcs && !selectedRepo) {
      setValidationError('Select a repository')
      return
    }

    onSubmit({
      branchName: branchName.trim(),
      useVcs,
      selectedVcsConnectionId,
      selectedRepo,
      selectedBranch,
      directory: directory.trim(),
      pathFilter: pathFilter.trim(),
    })
  }

  return (
    <Modal
      heading="Edit branch"
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save changes',
        onClick: handleSubmit,
        disabled: isSubmitting || !branchName.trim(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: onCancel,
        disabled: isSubmitting,
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {displayError && (
          <Banner theme="error" className="mb-4">
            {displayError}
          </Banner>
        )}

        <Input
          id="branch-name"
          type="text"
          value={branchName}
          onChange={(e) => setBranchName(e.target.value)}
          placeholder="Enter branch name"
          disabled={isSubmitting}
          autoFocus
          labelProps={{ labelText: 'Branch name' }}
        />

        <CheckboxInput
          id="use-vcs"
          checked={useVcs}
          onChange={(e) => {
            setUseVcs(e.target.checked)
            if (!e.target.checked) setValidationError(null)
          }}
          disabled={isSubmitting}
          labelProps={{ labelText: 'Connect to git repository' }}
        />

        {useVcs && (
          <>
            {vcsConnections.length === 0 ? (
              <Banner theme="warn">
                No VCS connections found. Connect your GitHub account first.
              </Banner>
            ) : (
              <>
                {vcsConnections.length > 1 && (
                  <Select
                    id="vcs-connection"
                    value={selectedVcsConnectionId}
                    onChange={(e) => onVcsConnectionChange(e.target.value)}
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
                  <>
                    <div className="flex flex-col gap-1">
                      <Skeleton width="80px" height="14px" />
                      <Skeleton height="36px" />
                    </div>
                    <div className="flex flex-col gap-1">
                      <Skeleton width="80px" height="14px" />
                      <Skeleton height="36px" />
                    </div>
                  </>
                ) : reposError ? (
                  <Banner theme="error">Failed to load repositories</Banner>
                ) : repos.length === 0 ? (
                  <Banner theme="warn">
                    No connected repositories found. Update your GitHub connection
                    to grant access to repositories.
                  </Banner>
                ) : (
                  <Select
                    id="repo-select"
                    value={selectedRepo?.full_name || ''}
                    onChange={(e) => {
                      const r = repos.find((r) => r.full_name === e.target.value)
                      onRepoChange(r || null)
                    }}
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
                    <div className="flex flex-col gap-1">
                      <Skeleton width="80px" height="14px" />
                      <Skeleton height="36px" />
                    </div>
                  ) : branchesError ? (
                    <Input
                      id="branch-select"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git branch' }}
                    />
                  ) : branches.length === 0 && selectedRepo ? (
                    <Input
                      id="branch-select"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git branch' }}
                      helperText="No branches found. Enter branch name manually."
                    />
                  ) : branches.length > 0 ? (
                    <Select
                      id="branch-select"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
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
                      id="branch-select"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git branch' }}
                    />
                  ))}

                <Input
                  id="directory"
                  type="text"
                  value={directory}
                  onChange={(e) => setDirectory(e.target.value)}
                  placeholder="."
                  disabled={isSubmitting}
                  labelProps={{ labelText: 'Directory' }}
                  helperText='Path to your application config (use "." for root)'
                />

                <Input
                  id="path-filter"
                  type="text"
                  value={pathFilter}
                  onChange={(e) => setPathFilter(e.target.value)}
                  placeholder="^(src/|config/).*"
                  disabled={isSubmitting}
                  labelProps={{ labelText: 'Path filter (optional)' }}
                  helperText="Regex pattern to filter which file changes trigger workflow runs"
                />
              </>
            )}
          </>
        )}
      </div>
    </Modal>
  )
}
