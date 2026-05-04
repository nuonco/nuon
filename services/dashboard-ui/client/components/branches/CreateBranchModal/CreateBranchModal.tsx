import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import type {
  TCreateAppBranchRequest,
  TVCSConnectionRepo,
  TVCSBranch,
  TVCSConnection,
} from '@/types'

interface ICreateBranchModal extends Omit<IModal, 'onSubmit'> {
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
  onSubmit: (
    body: TCreateAppBranchRequest & {
      vcs_connection_id?: string
      connected_github_vcs_config?: {
        repo: string
        branch: string
        directory: string
        path_filter?: string
      }
      public_git_vcs_config?: {
        repo: string
        branch: string
        directory: string
        path_filter?: string
      }
    }
  ) => void
  onCancel: () => void
}

export const CreateBranchModal = ({
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
  onSubmit,
  onCancel,
  ...props
}: ICreateBranchModal) => {
  const [name, setName] = useState('')
  const [useVcs, setUseVcs] = useState(true)
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [validationError, setValidationError] = useState<string | null>(null)

  const handleSubmit = () => {
    setValidationError(null)

    if (!name.trim()) {
      setValidationError('Branch name is required')
      return
    }

    if (useVcs) {
      if (!selectedRepo) {
        setValidationError('Repository is required when using VCS')
        return
      }
      if (!selectedBranch) {
        setValidationError('Git branch is required when using VCS')
        return
      }
    }

    const body: TCreateAppBranchRequest & {
      vcs_connection_id?: string
      connected_github_vcs_config?: {
        repo: string
        branch: string
        directory: string
        path_filter?: string
      }
      public_git_vcs_config?: {
        repo: string
        branch: string
        directory: string
        path_filter?: string
      }
    } = { name: name.trim() }

    if (useVcs && selectedRepo) {
      if (selectedRepo.private) {
        body.vcs_connection_id = selectedVcsConnectionId
        body.connected_github_vcs_config = {
          repo: selectedRepo.full_name,
          branch: selectedBranch,
          directory: directory.trim(),
        }
        if (pathFilter.trim()) {
          body.connected_github_vcs_config.path_filter = pathFilter.trim()
        }
      } else {
        body.public_git_vcs_config = {
          repo: selectedRepo.full_name,
          branch: selectedBranch,
          directory: directory.trim(),
        }
        if (pathFilter.trim()) {
          body.public_git_vcs_config.path_filter = pathFilter.trim()
        }
      }
    }

    onSubmit(body)
  }

  return (
    <Modal
      heading="Create app branch"
      primaryActionTrigger={{
        children: isSubmitting ? 'Creating...' : 'Create branch',
        disabled:
          isSubmitting ||
          (useVcs &&
            (loadingRepos ||
              loadingBranches ||
              vcsConnections.length === 0 ||
              !selectedRepo ||
              !selectedBranch)),
        onClick: handleSubmit,
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
        {validationError && <Banner theme="error">{validationError}</Banner>}

        <Input
          id="branch-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="production"
          required
          disabled={isSubmitting}
          labelProps={{ labelText: 'Branch name' }}
        />

        <CheckboxInput
          id="use-vcs"
          checked={useVcs}
          onChange={(e) => {
            setUseVcs(e.target.checked)
            if (!e.target.checked && validationError?.includes('VCS')) {
              setValidationError(null)
            }
          }}
          disabled={isSubmitting}
          labelProps={{ labelText: 'Connect to git repository' }}
        />

        {useVcs && (
          <>
            {vcsConnections.length === 0 ? (
              <Banner theme="warn">
                No VCS connections found. Please connect your GitHub account
                first.
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
                        conn.github_account_name ||
                        conn.github_install_id ||
                        conn.id,
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
                    No connected repositories found. Please update your GitHub
                    connection to grant access to repositories.
                  </Banner>
                ) : (
                  <Select
                    id="repo"
                    value={selectedRepo?.full_name || ''}
                    onChange={(e) => {
                      const repo = repos.find(
                        (r) => r.full_name === e.target.value
                      )
                      onRepoChange(repo || null)
                    }}
                    required={useVcs}
                    disabled={
                      isSubmitting || loadingRepos || loadingBranches || !useVcs
                    }
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
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isSubmitting || !useVcs}
                      labelProps={{ labelText: 'Git branch' }}
                    />
                  ) : branches.length === 0 && selectedRepo ? (
                    <Input
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isSubmitting || !useVcs}
                      labelProps={{ labelText: 'Git branch' }}
                      helperText="No branches found. Enter branch name manually."
                    />
                  ) : branches.length > 0 ? (
                    <Select
                      id="git-branch"
                      value={selectedBranch}
                      onChange={(e) => onBranchChange(e.target.value)}
                      required={useVcs}
                      disabled={isSubmitting || loadingBranches || !useVcs}
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
                      required={useVcs}
                      disabled={isSubmitting || !useVcs}
                      labelProps={{ labelText: 'Git branch' }}
                    />
                  ))}

                <Input
                  id="directory"
                  type="text"
                  value={directory}
                  onChange={(e) => setDirectory(e.target.value)}
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
