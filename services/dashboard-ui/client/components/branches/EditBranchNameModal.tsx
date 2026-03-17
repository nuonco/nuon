import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Banner } from '@/components/common/Banner'
import { Button, type IButtonAsButton } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Input } from '@/components/common/form/Input'
import { Select } from '@/components/common/form/Select'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Toast } from '@/components/surfaces/Toast'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useToast } from '@/hooks/use-toast'
import {
  createBranchConfig,
  updateBranch,
  getVCSConnectionRepos,
  type TVCSConnectionRepo,
} from '@/lib'
import {
  getConnectionBranches,
  type Branch,
} from '@/lib/ctl-api/vcs/get-connection-branches'
import type { TAppBranch, TAppBranchConfig } from '@/types'

interface IEditBranchNameModal extends IModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  onSuccess?: () => void
}

export const EditBranchNameModal = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: IEditBranchNameModal) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const vcsConnections = org?.vcs_connections || []

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
  const [branches, setBranches] = useState<Branch[]>([])
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
            orgId: org.id,
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
  }, [selectedVcsConnectionId, org.id, useVcs])

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
            org.id,
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
  }, [selectedRepo, selectedVcsConnectionId, org.id, useVcs])

  const formatError = (err: any): string => {
    if (!err) return 'An error occurred'
    if (typeof err === 'string') return err
    return (
      err.user_error && typeof err.user_error === 'string'
        ? err.user_error
        : err.error || err.description || err.message || 'An error occurred'
    )
  }

  const { mutate: handleSave, isPending: isSubmitting } = useMutation({
    mutationFn: async () => {
      if (!branchName.trim()) {
        throw new Error('Branch name cannot be empty')
      }

      if (useVcs && !selectedRepo) {
        throw new Error('Please select a repository')
      }

      if (branchName !== branch.name) {
        try {
          await updateBranch({
            appId: app.id,
            branchId: branch.id || '',
            orgId: org.id,
            request: { name: branchName },
          })
        } catch (err) {
          throw new Error(formatError(err))
        }
      }

      const request: any = {}

      if (useVcs && selectedRepo) {
        if (selectedRepo.private) {
          request.connected_github_vcs_config = {
            vcs_connection_id: selectedVcsConnectionId,
            repo: selectedRepo.full_name,
            branch: selectedBranch,
            directory: directory.trim(),
            path_filter: pathFilter.trim() || undefined,
          }
        } else {
          request.public_git_vcs_config = {
            repo: selectedRepo.full_name,
            branch: selectedBranch,
            directory: directory.trim(),
            path_filter: pathFilter.trim() || undefined,
          }
        }
      }

      if (currentConfig?.install_groups && currentConfig.install_groups.length > 0) {
        request.install_groups = currentConfig.install_groups.map((g, idx) => ({
          name: g.name,
          install_ids: g.install_ids || [],
          order: g.order ?? idx,
          max_parallel: g.max_parallel || 1,
          requires_approval: g.requires_approval || false,
          rollback_on_failure: g.rollback_on_failure || false,
        }))
      }

      const hasVCS = request.connected_github_vcs_config || request.public_git_vcs_config
      const hasGroups = (request.install_groups?.length ?? 0) > 0

      if (hasVCS || hasGroups) {
        try {
          await createBranchConfig({
            appId: app.id,
            branchId: branch.id || '',
            orgId: org.id,
            request,
          })
        } catch (err) {
          throw new Error(formatError(err))
        }
      }
    },
    onSuccess: () => {
      addToast(
        <Toast heading="Branch updated successfully" theme="success">
          <Text>Updated branch: {branchName}</Text>
        </Toast>
      )
      setValidationError(null)
      onSuccess?.()
      removeModal(props.modalId)
    },
    onError: (error: Error) => {
      const msg = error?.message || 'An error occurred'
      setValidationError(msg)
      addToast(
        <Toast heading="Branch update failed" theme="error">
          <Text>{msg}</Text>
        </Toast>
      )
    },
  })

  return (
    <Modal
      heading="Edit Branch"
      size="3/4"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save Changes',
        onClick: () => handleSave(),
        disabled: isSubmitting || !branchName.trim(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: () => removeModal(props.modalId),
        disabled: isSubmitting,
      }}
      {...props}
    >
      <div className="flex flex-col gap-4">
        {validationError && (
          <Banner theme="error" className="mb-4">
            {validationError}
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
              <Banner theme="warning">
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
                  <Banner theme="warning">
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

export const EditBranchButton = ({
  branch,
  currentConfig,
  onSuccess,
  ...props
}: { branch: TAppBranch; currentConfig?: TAppBranchConfig; onSuccess?: () => void } & Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <EditBranchNameModal branch={branch} currentConfig={currentConfig} onSuccess={onSuccess} />
  return (
    <Button variant="secondary" size="sm" onClick={() => addModal(modal)} {...props}>
      <Icon variant="Edit" size={16} />
      Edit
    </Button>
  )
}
