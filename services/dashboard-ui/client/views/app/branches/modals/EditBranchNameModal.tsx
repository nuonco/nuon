import { useEffect, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { Modal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { Toast } from '@/components/surfaces/Toast'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
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

interface IEditBranchNameModal {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  isVisible: boolean
  onClose: () => void
}

export const EditBranchNameModal = ({
  branch,
  currentConfig,
  isVisible,
  onClose,
}: IEditBranchNameModal) => {
  const { app } = useApp()
  const { org } = useOrg()
  const { refresh } = useBranch()
  const { addToast } = useToast()

  const [branchName, setBranchName] = useState(branch.name || '')

  // "Connect to Git Repository" checkbox — mirrors CreateBranchModal
  const [useVcs, setUseVcs] = useState(
    !!(currentConfig?.connected_github_vcs_config || currentConfig?.public_git_vcs_config)
  )

  // VCS selector state — mirrors CreateBranchModal
  const [selectedVcsConnectionId, setSelectedVcsConnectionId] = useState('')
  const [repos, setRepos] = useState<TVCSConnectionRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<TVCSConnectionRepo | null>(null)
  const [branches, setBranches] = useState<Branch[]>([])
  const [selectedBranch, setSelectedBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)
  const [reposError, setReposError] = useState<string | null>(null)
  const [branchesError, setBranchesError] = useState<string | null>(null)

  const [validationError, setValidationError] = useState<string | null>(null)

  const vcsConnections = org?.vcs_connections || []

  // Initialise state from currentConfig when modal opens
  useEffect(() => {
    if (!isVisible) return

    const hasExistingVcs = !!(
      currentConfig?.connected_github_vcs_config ||
      currentConfig?.public_git_vcs_config
    )
    setUseVcs(hasExistingVcs)

    const existingConnectionId =
      currentConfig?.connected_github_vcs_config?.vcs_connection_id || ''
    const existingDir =
      currentConfig?.connected_github_vcs_config?.directory ||
      currentConfig?.public_git_vcs_config?.directory ||
      '.'
    const existingPathFilter =
      currentConfig?.connected_github_vcs_config?.path_filter ||
      currentConfig?.public_git_vcs_config?.path_filter ||
      ''

    setDirectory(existingDir)
    setPathFilter(existingPathFilter)
    setBranchName(branch.name || '')

    if (existingConnectionId) {
      setSelectedVcsConnectionId(existingConnectionId)
    } else if (vcsConnections.length > 0) {
      setSelectedVcsConnectionId(vcsConnections[0].id)
    }
  }, [isVisible, currentConfig])

  // Seed connection when vcsConnections loads (if not already set)
  useEffect(() => {
    if (vcsConnections.length > 0 && !selectedVcsConnectionId && isVisible) {
      setSelectedVcsConnectionId(vcsConnections[0].id)
    }
  }, [vcsConnections, isVisible])

  // Fetch repos when connection changes
  useEffect(() => {
    if (!selectedVcsConnectionId || !isVisible || !useVcs) {
      setRepos([])
      setSelectedRepo(null)
      setReposError(null)
      return
    }

    const existingRepo =
      currentConfig?.connected_github_vcs_config?.repo ||
      currentConfig?.public_git_vcs_config?.repo ||
      ''

    const fetchRepos = async () => {
      setLoadingRepos(true)
      setReposError(null)
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
          const match = existingRepo
            ? sorted.find((r) => r.full_name === existingRepo)
            : null
          setSelectedRepo(match ?? (sorted.length > 0 ? sorted[0] : null))
        } else {
          setRepos([])
          setSelectedRepo(null)
        }
      } catch {
        setReposError('Failed to load repositories. Please check your VCS connection.')
        setRepos([])
        setSelectedRepo(null)
      } finally {
        setLoadingRepos(false)
      }
    }
    fetchRepos()
  }, [selectedVcsConnectionId, isVisible, useVcs, org.id])

  // Fetch branches when repo changes
  useEffect(() => {
    if (!selectedRepo || !selectedVcsConnectionId || !isVisible || !useVcs) {
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
      setBranchesError(null)
      try {
        const fetched = await getConnectionBranches(
          org.id,
          selectedVcsConnectionId,
          owner,
          repoName
        )
        setBranches(fetched)
        const match = fetched.find((b) => b.name === existingBranch)
        if (match) {
          setSelectedBranch(match.name)
        } else {
          const main = fetched.find((b) => b.name === 'main')
          setSelectedBranch(main ? 'main' : fetched.length > 0 ? fetched[0].name : 'main')
        }
      } catch {
        setBranchesError('Failed to load branches. Please try again.')
        setBranches([])
      } finally {
        setLoadingBranches(false)
      }
    }
    fetchBranches()
  }, [selectedRepo, selectedVcsConnectionId, isVisible, useVcs, org.id])

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

      // Step 1: Update branch name if changed
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

      // Step 2: Build and create new config
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

      // Preserve install groups from currentConfig
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
      refresh()
      onClose()
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
      isVisible={isVisible}
      onClose={onClose}
      heading="Edit Branch"
      size="3/4"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save Changes',
        onClick: () => handleSave(),
        disabled: isSubmitting || !branchName.trim(),
      }}
    >
      {validationError && (
        <Banner theme="error" className="mb-4">
          {validationError}
        </Banner>
      )}

      <div className="flex flex-col gap-4">
        {/* Branch Name */}
        <div className="flex flex-col gap-2">
          <label htmlFor="branch-name" className="text-sm font-medium">
            Branch Name
          </label>
          <input
            id="branch-name"
            type="text"
            value={branchName}
            onChange={(e) => setBranchName(e.target.value)}
            className="w-full p-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            placeholder="Enter branch name"
            disabled={isSubmitting}
            autoFocus
          />
        </div>

        {/* VCS Section */}
        <div className="border-t border-gray-200 dark:border-gray-700 pt-4 flex flex-col gap-4">
          <div className="flex items-center gap-2">
            <input
              id="use-vcs"
              type="checkbox"
              checked={useVcs}
              onChange={(e) => {
                setUseVcs(e.target.checked)
                if (!e.target.checked) setValidationError(null)
              }}
              disabled={isSubmitting}
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
            />
            <label htmlFor="use-vcs" className="text-sm font-medium">
              Connect to Git Repository
            </label>
          </div>

          {useVcs && (
            <>
              {vcsConnections.length === 0 ? (
                <Banner theme="warning">
                  No VCS connections found. Please connect your GitHub account first.
                </Banner>
              ) : (
                <>
                  {vcsConnections.length > 1 && (
                    <div className="flex flex-col gap-2">
                      <label htmlFor="vcs-connection" className="text-sm font-medium">
                        VCS Connection
                      </label>
                      <select
                        id="vcs-connection"
                        value={selectedVcsConnectionId}
                        onChange={(e) => setSelectedVcsConnectionId(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        disabled={isSubmitting || loadingRepos}
                      >
                        {vcsConnections.map((conn) => (
                          <option key={conn.id} value={conn.id}>
                            {conn.github_account_name || conn.github_install_id || conn.id}
                          </option>
                        ))}
                      </select>
                    </div>
                  )}

                  {reposError && <Banner theme="error">{reposError}</Banner>}

                  <div className="flex flex-col gap-2">
                    <label htmlFor="repo-select" className="text-sm font-medium">
                      Repository
                    </label>
                    {loadingRepos ? (
                      <div className="px-3 py-2 text-sm text-gray-500">
                        Loading repositories...
                      </div>
                    ) : reposError ? (
                      <div className="px-3 py-2 text-sm text-red-500">
                        Failed to load repositories
                      </div>
                    ) : repos.length === 0 ? (
                      <Banner theme="warning">
                        No connected repositories found. Please update your GitHub
                        connection to grant access to repositories.
                      </Banner>
                    ) : (
                      <select
                        id="repo-select"
                        value={selectedRepo?.full_name || ''}
                        onChange={(e) => {
                          const r = repos.find((r) => r.full_name === e.target.value)
                          setSelectedRepo(r || null)
                        }}
                        className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        disabled={isSubmitting || loadingRepos || loadingBranches}
                      >
                        {repos.map((r) => (
                          <option key={r.id || r.full_name} value={r.full_name}>
                            {r.full_name}{r.private ? ' 🔒' : ''}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>

                  {branchesError && <Banner theme="error">{branchesError}</Banner>}

                  <div className="flex flex-col gap-2">
                    <label htmlFor="branch-select" className="text-sm font-medium">
                      Git Branch
                    </label>
                    {loadingBranches ? (
                      <div className="px-3 py-2 text-sm text-gray-500">
                        Loading branches...
                      </div>
                    ) : branches.length > 0 ? (
                      <select
                        id="branch-select"
                        value={selectedBranch}
                        onChange={(e) => setSelectedBranch(e.target.value)}
                        className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        disabled={isSubmitting || loadingBranches}
                      >
                        {branches.map((b) => (
                          <option key={b.name} value={b.name}>
                            {b.name}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <Input
                        id="branch-select"
                        type="text"
                        value={selectedBranch}
                        onChange={(e) => setSelectedBranch(e.target.value)}
                        placeholder="main"
                        disabled={isSubmitting}
                      />
                    )}
                  </div>

                  <div className="flex flex-col gap-2">
                    <label htmlFor="directory" className="text-sm font-medium">
                      Directory
                    </label>
                    <Input
                      id="directory"
                      type="text"
                      value={directory}
                      onChange={(e) => setDirectory(e.target.value)}
                      placeholder="."
                      disabled={isSubmitting}
                    />
                    <p className="text-xs text-gray-500">
                      Path to your application config (use &quot;.&quot; for root)
                    </p>
                  </div>

                  <div className="flex flex-col gap-2">
                    <label htmlFor="path-filter" className="text-sm font-medium">
                      Path Filter (Optional)
                    </label>
                    <Input
                      id="path-filter"
                      type="text"
                      value={pathFilter}
                      onChange={(e) => setPathFilter(e.target.value)}
                      placeholder="^(src/|config/).*"
                      disabled={isSubmitting}
                    />
                    <p className="text-xs text-gray-500">
                      Regex pattern to filter which file changes trigger workflow runs
                    </p>
                  </div>
                </>
              )}
            </>
          )}
        </div>
      </div>
    </Modal>
  )
}
