'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Modal } from '@/components/surfaces/Modal'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import {
  createBranchConfig,
  getAppInstalls,
  getVCSConnectionRepos,
} from '@/lib'
import {
  getConnectionBranches,
  type Branch,
} from '@/lib/ctl-api/vcs/get-connection-branches'
import type {
  TAppBranch,
  TAppBranchConfig,
  TInstall,
  TVCSConnection,
  TVCSConnectionRepo,
} from '@/types'

interface IInstallGroup {
  id: string
  name: string
  install_ids: string[]
  order: number
  max_parallel: number
  requires_approval: boolean
  rollback_on_failure: boolean
}

interface INewBranchConfigPanel {
  branch: TAppBranch
  currentConfig?: TAppBranchConfig
  vcsConnections: TVCSConnection[]
  orgId: string
  appId: string
  isVisible: boolean
  onClose: () => void
}

export const NewBranchConfigPanel = ({
  branch,
  currentConfig,
  vcsConnections,
  orgId,
  appId,
  isVisible,
  onClose,
}: INewBranchConfigPanel) => {
  const router = useRouter()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // VCS Configuration
  const [selectedVcsConnectionId, setSelectedVcsConnectionId] = useState('')
  const [repos, setRepos] = useState<TVCSConnectionRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState('')
  const [branches, setBranches] = useState<Branch[]>([])
  const [selectedBranch, setSelectedBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)

  // Install Groups
  const [installGroups, setInstallGroups] = useState<IInstallGroup[]>([])
  const [availableInstalls, setAvailableInstalls] = useState<TInstall[]>([])
  const [loadingInstalls, setLoadingInstalls] = useState(false)

  const nextConfigNumber = (currentConfig?.config_number || 0) + 1

  // Initialize from current config
  useEffect(() => {
    if (isVisible && currentConfig) {
      // Load VCS config from either connected GitHub or public git
      if (currentConfig.connected_github_vcs_config) {
        const vcs = currentConfig.connected_github_vcs_config
        setSelectedVcsConnectionId(vcs.vcs_connection_id || '')
        setSelectedRepo(vcs.repo || '')
        setSelectedBranch(vcs.branch || 'main')
        setDirectory(vcs.directory || '.')
        setPathFilter(vcs.path_filter || '')
      } else if (currentConfig.public_git_vcs_config) {
        const vcs = currentConfig.public_git_vcs_config
        setSelectedRepo(vcs.repo || '')
        setSelectedBranch(vcs.branch || 'main')
        setDirectory(vcs.directory || '.')
        setPathFilter(vcs.path_filter || '')
      }

      if (currentConfig.install_groups) {
        setInstallGroups(
          currentConfig.install_groups.map((group, idx) => ({
            id: `group-${idx}`,
            name: group.name || '',
            install_ids: group.install_ids || [],
            order: group.order || idx,
            max_parallel: group.max_parallel || 1,
            requires_approval: group.requires_approval || false,
            rollback_on_failure: group.rollback_on_failure || false,
          }))
        )
      }
    } else if (isVisible && vcsConnections.length > 0) {
      setSelectedVcsConnectionId(vcsConnections[0].id || '')
    }
  }, [isVisible, currentConfig, vcsConnections])

  // Fetch available installs
  useEffect(() => {
    if (!isVisible) return

    const fetchInstalls = async () => {
      setLoadingInstalls(true)
      const { data, error: installsError } = await getAppInstalls({
        appId,
        orgId,
        limit: 100,
      })

      if (installsError) {
        console.error('Failed to fetch installs:', installsError)
      } else {
        setAvailableInstalls(data || [])
      }
      setLoadingInstalls(false)
    }

    fetchInstalls()
  }, [isVisible, appId, orgId])

  // Fetch repos when VCS connection changes
  useEffect(() => {
    if (!selectedVcsConnectionId || !isVisible) {
      setRepos([])
      return
    }

    const fetchRepos = async () => {
      setLoadingRepos(true)
      try {
        const response = await getVCSConnectionRepos({
          orgId,
          connectionId: selectedVcsConnectionId,
        })
        if (
          response.data?.repositories &&
          Array.isArray(response.data.repositories)
        ) {
          setRepos(response.data.repositories)
          if (response.data.repositories.length > 0 && !selectedRepo) {
            setSelectedRepo(response.data.repositories[0].full_name)
          }
        } else {
          setRepos([])
        }
      } catch (err) {
        console.error('Failed to fetch repos:', err)
        setRepos([])
      } finally {
        setLoadingRepos(false)
      }
    }

    fetchRepos()
  }, [selectedVcsConnectionId, isVisible, orgId, selectedRepo])

  // Fetch branches when repo changes
  useEffect(() => {
    if (!selectedRepo || !selectedVcsConnectionId || !isVisible) {
      setBranches([])
      return
    }

    const [owner, repo] = selectedRepo.split('/')
    if (!owner || !repo) return

    const fetchBranches = async () => {
      setLoadingBranches(true)
      try {
        const response = await getConnectionBranches(
          orgId,
          selectedVcsConnectionId,
          owner,
          repo
        )
        if (response.data) {
          setBranches(response.data)
          if (response.data.length > 0 && !selectedBranch) {
            const mainBranch = response.data.find((b) => b.name === 'main')
            setSelectedBranch(mainBranch ? 'main' : response.data[0].name)
          }
        }
      } catch (err) {
        console.error('Failed to fetch branches:', err)
      } finally {
        setLoadingBranches(false)
      }
    }

    fetchBranches()
  }, [selectedRepo, selectedVcsConnectionId, isVisible, orgId, selectedBranch])

  const addInstallGroup = () => {
    const newGroup: IInstallGroup = {
      id: `group-${Date.now()}`,
      name: '',
      install_ids: [],
      order: installGroups.length,
      max_parallel: 1,
      requires_approval: false,
      rollback_on_failure: false,
    }
    setInstallGroups([...installGroups, newGroup])
  }

  const removeInstallGroup = (id: string) => {
    setInstallGroups(installGroups.filter((g) => g.id !== id))
  }

  const updateInstallGroup = (id: string, updates: Partial<IInstallGroup>) => {
    setInstallGroups(
      installGroups.map((g) => (g.id === id ? { ...g, ...updates } : g))
    )
  }

  const handleSubmit = async () => {
    setError(null)

    if (!selectedRepo || !selectedBranch) {
      setError('Repository and branch are required')
      return
    }

    // Validate install groups only if any exist
    if (installGroups.length > 0) {
      if (installGroups.some((g) => !g.name || g.install_ids.length === 0)) {
        setError('All install groups must have a name and at least one install')
        return
      }
    }

    setIsSubmitting(true)

    // Determine if using connected GitHub or public git based on VCS connection
    const repoData = repos.find((r) => r.full_name === selectedRepo)
    const isPrivateRepo = repoData?.private && selectedVcsConnectionId

    const requestBody: any = {}

    // Always include VCS config
    if (isPrivateRepo) {
      requestBody.connected_github_vcs_config = {
        vcs_connection_id: selectedVcsConnectionId,
        repo: selectedRepo,
        branch: selectedBranch,
        directory: directory.trim(),
        path_filter: pathFilter.trim() || undefined,
      }
    } else {
      requestBody.public_git_vcs_config = {
        repo: selectedRepo,
        branch: selectedBranch,
        directory: directory.trim(),
        path_filter: pathFilter.trim() || undefined,
      }
    }

    // Include install groups only if any exist
    if (installGroups.length > 0) {
      requestBody.install_groups = installGroups.map((g) => ({
        name: g.name,
        install_ids: g.install_ids,
        order: g.order,
        max_parallel: g.max_parallel,
        requires_approval: g.requires_approval,
        rollback_on_failure: g.rollback_on_failure,
      }))
    }

    const { error: createError } = await createBranchConfig({
      appId,
      branchId: branch.id || '',
      orgId,
      request: requestBody,
    })

    setIsSubmitting(false)

    if (createError) {
      setError(
        typeof createError === 'string'
          ? createError
          : createError.user_error ||
              createError.error ||
              createError.description ||
              'Failed to create branch configuration'
      )
    } else {
      router.refresh()
      onClose()
    }
  }

  return (
    <Modal
      isVisible={isVisible}
      onClose={onClose}
      heading={
        <div>
          <Text variant="h3" weight="strong">
            Edit Configuration
          </Text>
          <Text variant="subtext" theme="neutral">
            Editing will create version {nextConfigNumber}
          </Text>
        </div>
      }
      size="3/4"
      primaryActionTrigger={{
        children: isSubmitting
          ? 'Creating...'
          : `Create Configuration v${nextConfigNumber}`,
        onClick: handleSubmit,
        disabled: isSubmitting || loadingRepos || loadingBranches,
      }}
    >
      {error && (
        <Banner theme="error" className="mb-4">
          {error}
        </Banner>
      )}

      <div className="space-y-6">
        {/* VCS Configuration */}
        <div className="space-y-4">
          <Text variant="h4" weight="strong">
            VCS Configuration
          </Text>

          {vcsConnections.length === 0 ? (
            <Banner theme="warning">
              No VCS connections found. Please connect your GitHub account
              first.
            </Banner>
          ) : (
            <>
              {vcsConnections.length > 1 && (
                <div>
                  <label
                    htmlFor="vcs-connection"
                    className="block text-sm font-medium mb-2"
                  >
                    VCS Connection
                  </label>
                  <select
                    id="vcs-connection"
                    value={selectedVcsConnectionId}
                    onChange={(e) => setSelectedVcsConnectionId(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={isSubmitting}
                  >
                    {vcsConnections.map((conn) => (
                      <option key={conn.id} value={conn.id}>
                        {conn.github_account_name || conn.id}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <div>
                <label
                  htmlFor="repo"
                  className="block text-sm font-medium mb-2"
                >
                  Repository
                </label>
                {loadingRepos ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    Loading repositories...
                  </div>
                ) : repos.length === 0 ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    No repositories found
                  </div>
                ) : (
                  <select
                    id="repo"
                    value={selectedRepo}
                    onChange={(e) => setSelectedRepo(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={isSubmitting}
                  >
                    {Array.isArray(repos) &&
                      repos.map((repo) => (
                        <option
                          key={repo.id || repo.full_name}
                          value={repo.full_name}
                        >
                          {repo.full_name}
                          {repo.private ? ' 🔒' : ''}
                        </option>
                      ))}
                  </select>
                )}
              </div>

              <div>
                <label
                  htmlFor="git-branch"
                  className="block text-sm font-medium mb-2"
                >
                  Branch
                </label>
                {loadingBranches ? (
                  <div className="px-3 py-2 text-sm text-gray-500">
                    Loading branches...
                  </div>
                ) : branches.length === 0 ? (
                  <Input
                    id="git-branch"
                    type="text"
                    value={selectedBranch}
                    onChange={(e) => setSelectedBranch(e.target.value)}
                    placeholder="main"
                    disabled={isSubmitting}
                  />
                ) : (
                  <select
                    id="git-branch"
                    value={selectedBranch}
                    onChange={(e) => setSelectedBranch(e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    disabled={isSubmitting}
                  >
                    {branches.map((branch) => (
                      <option key={branch.name} value={branch.name}>
                        {branch.name}
                      </option>
                    ))}
                  </select>
                )}
              </div>

              <div>
                <label
                  htmlFor="directory"
                  className="block text-sm font-medium mb-2"
                >
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
                <Text variant="subtext" theme="neutral" className="mt-1">
                  Path to your application config (use &quot;.&quot; for root)
                </Text>
              </div>

              <div>
                <label
                  htmlFor="path-filter"
                  className="block text-sm font-medium mb-2"
                >
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
                <Text variant="subtext" theme="neutral" className="mt-1">
                  Regex pattern to filter which file changes trigger workflow
                  runs
                </Text>
              </div>
            </>
          )}
        </div>

        {/* Install Groups */}
        <div className="space-y-4 border-t pt-6">
          <div className="flex items-center justify-between">
            <Text variant="h4" weight="strong">
              Install Groups
            </Text>
            <Button
              onClick={addInstallGroup}
              variant="secondary"
              size="sm"
              disabled={isSubmitting || loadingInstalls}
            >
              <Icon variant="Plus" size={16} />
              Add Group
            </Button>
          </div>

          {loadingInstalls ? (
            <div className="px-3 py-2 text-sm text-gray-500">
              Loading installs...
            </div>
          ) : availableInstalls.length === 0 ? (
            <Banner theme="info">
              No installs found for this app. Create installs first to configure
              deployment groups.
            </Banner>
          ) : (
            <div className="space-y-4">
              {installGroups.map((group, idx) => (
                <div
                  key={group.id}
                  className="p-4 border border-gray-300 dark:border-gray-600 rounded-lg space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <Text variant="base" weight="strong">
                      Group {idx + 1}
                    </Text>
                    <Button
                      onClick={() => removeInstallGroup(group.id)}
                      variant="ghost"
                      size="sm"
                      disabled={isSubmitting}
                    >
                      <Icon variant="Trash" size={16} />
                      Remove
                    </Button>
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Name
                    </label>
                    <Input
                      type="text"
                      value={group.name}
                      onChange={(e) =>
                        updateInstallGroup(group.id, { name: e.target.value })
                      }
                      placeholder="e.g., staging, production"
                      disabled={isSubmitting}
                    />
                  </div>

                  <div>
                    <label className="block text-sm font-medium mb-2">
                      Installs
                    </label>
                    <select
                      multiple
                      value={group.install_ids}
                      onChange={(e) => {
                        const selected = Array.from(
                          e.target.selectedOptions,
                          (option) => option.value
                        )
                        updateInstallGroup(group.id, { install_ids: selected })
                      }}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[100px]"
                      disabled={isSubmitting}
                    >
                      {availableInstalls.map((install) => (
                        <option key={install.id} value={install.id}>
                          {install.name}
                        </option>
                      ))}
                    </select>
                    <Text variant="subtext" theme="neutral" className="mt-1">
                      Hold Ctrl/Cmd to select multiple installs
                    </Text>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <label className="block text-sm font-medium mb-2">
                        Max Parallel
                      </label>
                      <Input
                        type="number"
                        min="1"
                        value={group.max_parallel}
                        onChange={(e) =>
                          updateInstallGroup(group.id, {
                            max_parallel: parseInt(e.target.value) || 1,
                          })
                        }
                        disabled={isSubmitting}
                      />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <label className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={group.requires_approval}
                        onChange={(e) =>
                          updateInstallGroup(group.id, {
                            requires_approval: e.target.checked,
                          })
                        }
                        disabled={isSubmitting}
                        className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      />
                      <Text variant="base">Requires Approval</Text>
                    </label>

                    <label className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={group.rollback_on_failure}
                        onChange={(e) =>
                          updateInstallGroup(group.id, {
                            rollback_on_failure: e.target.checked,
                          })
                        }
                        disabled={isSubmitting}
                        className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                      />
                      <Text variant="base">Rollback on Failure</Text>
                    </label>
                  </div>
                </div>
              ))}

              {installGroups.length === 0 && (
                <Banner theme="info">
                  Click &quot;Add Group&quot; to create your first install group
                </Banner>
              )}
            </div>
          )}
        </div>
      </div>
    </Modal>
  )
}
