import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { getVCSConnectionRepos, getConnectionBranches } from '@/lib'
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
  orgId: string
  vcsConnections: TVCSConnection[]
  isSubmitting: boolean
  validationError: string | null
  onSubmit: (data: IEditBranchNameModalSubmitData) => void
  onCancel: () => void
}

export const EditBranchNameModal = ({
  branch,
  currentConfig,
  orgId,
  vcsConnections,
  isSubmitting,
  validationError: externalValidationError,
  onSubmit,
  onCancel,
  ...props
}: IEditBranchNameModal) => {
  const existingConnectionId =
    currentConfig?.connected_github_vcs_config?.vcs_connection_id || ''

  const [branchName, setBranchName] = useState(branch.name || '')
  const [useVcs, setUseVcs] = useState(
    !!(currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config)
  )
  const [selectedVcsConnectionId, setSelectedVcsConnectionId] = useState(
    existingConnectionId || vcsConnections[0]?.id || ''
  )
  const [repos, setRepos] = useState<TVCSConnectionRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<TVCSConnectionRepo | null>(null)
  const [branches, setBranches] = useState<TVCSBranch[]>([])
  const [selectedBranch, setSelectedBranch] = useState(
    currentConfig?.connected_github_vcs_config?.branch ||
      currentConfig?.public_git_vcs_config?.branch ||
      'main'
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
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)
  const [reposError, setReposError] = useState<string | null>(null)
  const [branchesError, setBranchesError] = useState<string | null>(null)
  const [validationError, setValidationError] = useState<string | null>(null)

  const displayError = externalValidationError || validationError

  useEffect(() => {
    if (!selectedVcsConnectionId || !useVcs) {
      setRepos([])
      setSelectedRepo(null)
      setReposError(null)
      return
    }

    setSelectedRepo(null)

    const existingRepo =
      currentConfig?.connected_github_vcs_config?.repo ||
      currentConfig?.public_git_vcs_config?.repo ||
      ''

    const fetchRepos = async () => {
      setLoadingRepos(true)
      setValidationError(null)
      setReposError(null)

      let lastErr: unknown
      for (let attempt = 0; attempt < 3; attempt++) {
        if (attempt > 0) await new Promise((r) => setTimeout(r, 1000 * attempt))
        try {
          const response = await getVCSConnectionRepos({
            orgId,
            connectionId: selectedVcsConnectionId,
          })

          if (response.repositories && Array.isArray(response.repositories)) {
            const sorted = [...response.repositories].sort((a, b) =>
              a.full_name.localeCompare(b.full_name)
            )
            setRepos(sorted)
            setReposError(null)
            const match = existingRepo
              ? sorted.find((r) => r.full_name === existingRepo)
              : null
            setSelectedRepo(match ?? (sorted.length > 0 ? sorted[0] : null))
          } else {
            setRepos([])
            setSelectedRepo(null)
          }
          setLoadingRepos(false)
          return
        } catch (err) {
          lastErr = err
        }
      }

      setReposError('Failed to load repositories. Please check your VCS connection.')
      setRepos([])
      setSelectedRepo(null)
      setLoadingRepos(false)
    }

    fetchRepos()
  }, [selectedVcsConnectionId, orgId, useVcs])

  useEffect(() => {
    if (!selectedRepo || !selectedVcsConnectionId || !useVcs) {
      setBranches([])
      setBranchesError(null)
      return
    }

    const existingBranch =
      currentConfig?.connected_github_vcs_config?.branch ||
      currentConfig?.public_git_vcs_config?.branch ||
      'main'

    const [owner, repoName] = selectedRepo.full_name.split('/')
    if (!owner || !repoName) return

    const fetchBranches = async () => {
      setLoadingBranches(true)
      setValidationError(null)
      setBranchesError(null)

      let lastErr: unknown
      for (let attempt = 0; attempt < 3; attempt++) {
        if (attempt > 0) await new Promise((r) => setTimeout(r, 1000 * attempt))
        try {
          const fetched = await getConnectionBranches(
            orgId,
            selectedVcsConnectionId,
            owner,
            repoName
          )
          setBranches(fetched)
          setBranchesError(null)
          const match = fetched.find((b) => b.name === existingBranch)
          if (match) {
            setSelectedBranch(match.name)
          } else {
            const main = fetched.find((b) => b.name === 'main')
            setSelectedBranch(main ? 'main' : fetched.length > 0 ? fetched[0].name : 'main')
          }
          setLoadingBranches(false)
          return
        } catch (err) {
          lastErr = err
        }
      }

      setBranchesError('Failed to load branches. Please try again.')
      setBranches([])
      setLoadingBranches(false)
    }

    fetchBranches()
  }, [selectedRepo, selectedVcsConnectionId, orgId, useVcs])

  const handleSubmit = () => {
    setValidationError(null)

    if (!branchName.trim()) {
      setValidationError('Branch name cannot be empty')
      return
    }

    if (useVcs && !selectedRepo) {
      setValidationError('Please select a repository')
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
      heading="Edit Branch"
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save Changes',
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
          labelProps={{ labelText: 'Branch Name' }}
        />

        <CheckboxInput
          id="use-vcs"
          checked={useVcs}
          onChange={(e) => {
            setUseVcs(e.target.checked)
            if (!e.target.checked) setValidationError(null)
          }}
          disabled={isSubmitting}
          labelProps={{ labelText: 'Connect to Git Repository' }}
        />

        {useVcs && (
          <>
            {vcsConnections.length === 0 ? (
              <Banner theme="warn">
                No VCS connections found. Please connect your GitHub account first.
              </Banner>
            ) : (
              <>
                {vcsConnections.length > 1 && (
                  <Select
                    id="vcs-connection"
                    value={selectedVcsConnectionId}
                    onChange={(e) => setSelectedVcsConnectionId(e.target.value)}
                    disabled={isSubmitting || loadingRepos}
                    options={vcsConnections.map((conn) => ({
                      value: conn.id,
                      label:
                        conn.github_account_name || conn.github_install_id || conn.id,
                    }))}
                    labelProps={{ labelText: 'VCS Connection' }}
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
                    No connected repositories found. Please update your GitHub connection
                    to grant access to repositories.
                  </Banner>
                ) : (
                  <Select
                    id="repo-select"
                    value={selectedRepo?.full_name || ''}
                    onChange={(e) => {
                      const r = repos.find((r) => r.full_name === e.target.value)
                      setSelectedRepo(r || null)
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
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git Branch' }}
                    />
                  ) : branches.length === 0 && selectedRepo ? (
                    <Input
                      id="branch-select"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git Branch' }}
                      helperText="No branches found. Enter branch name manually."
                    />
                  ) : branches.length > 0 ? (
                    <Select
                      id="branch-select"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      disabled={isSubmitting || loadingBranches}
                      options={branches.map((b) => ({
                        value: b.name,
                        label: b.name,
                      }))}
                      labelProps={{ labelText: 'Git Branch' }}
                      searchable
                    />
                  ) : (
                    <Input
                      id="branch-select"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      disabled={isSubmitting}
                      labelProps={{ labelText: 'Git Branch' }}
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
                  labelProps={{ labelText: 'Path Filter (Optional)' }}
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
