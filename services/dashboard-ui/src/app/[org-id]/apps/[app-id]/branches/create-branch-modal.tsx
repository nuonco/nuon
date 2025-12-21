'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Button } from '@/components/common/Button'
import { Input } from '@/components/common/form/Input'
import { Text } from '@/components/common/Text'
import { ModalBase } from '@/components/surfaces/Modal'
import { useOrg } from '@/hooks/use-org'
import { useServerAction } from '@/hooks/use-server-action'
import { useServerActionToast } from '@/hooks/use-server-action-toast'
import { getVCSConnectionRepos } from '@/lib'
import {
  getConnectionBranches,
  type Branch,
} from '@/lib/ctl-api/vcs/get-connection-branches'
import { createAppBranch } from './create-branch-action'
import type { TCreateAppBranchRequest, TVCSConnectionRepo } from '@/types'

interface ICreateBranchModal {
  appId: string
  orgId: string
  isOpen: boolean
  onClose: () => void
}

export const CreateBranchModal = ({
  appId,
  orgId,
  isOpen,
  onClose,
}: ICreateBranchModal) => {
  const router = useRouter()
  const path = usePathname()
  const { org } = useOrg()
  const [name, setName] = useState('')
  const [useVcs, setUseVcs] = useState(true)
  const [selectedVcsConnectionId, setSelectedVcsConnectionId] = useState('')
  const [repos, setRepos] = useState<TVCSConnectionRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<TVCSConnectionRepo | null>(
    null
  )
  const [branches, setBranches] = useState<Branch[]>([])
  const [selectedBranch, setSelectedBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [reposError, setReposError] = useState<string | null>(null)
  const [branchesError, setBranchesError] = useState<string | null>(null)

  // Helper function to extract error message string from various error formats
  const formatError = (err: any): string => {
    if (!err) return ''
    if (typeof err === 'string') return err
    return (
      err.user_error ||
      err.error ||
      err.description ||
      err.message ||
      'An error occurred'
    )
  }

  const { data, error, isLoading, execute } = useServerAction({
    action: createAppBranch,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>Failed to create app branch.</Text>
        <Text>
          {typeof error === 'string'
            ? error
            : error?.error || error?.description || 'Unknown error occurred.'}
        </Text>
      </>
    ),
    errorHeading: 'Branch creation failed',
    onSuccess: () => {
      if (data) {
        handleClose()
        router.push(`/${orgId}/apps/${appId}/branches/${data.id}`)
      }
    },
    successContent: <Text>Created app branch: {name}</Text>,
    successHeading: 'Branch created successfully',
  })

  const vcsConnections = org?.vcs_connections || []

  // Set first VCS connection as default
  useEffect(() => {
    if (
      isOpen &&
      vcsConnections.length > 0 &&
      !selectedVcsConnectionId &&
      useVcs
    ) {
      setSelectedVcsConnectionId(vcsConnections[0].id)
    }
  }, [isOpen, vcsConnections, selectedVcsConnectionId, useVcs])

  // Fetch repos when VCS connection changes
  useEffect(() => {
    if (!selectedVcsConnectionId || !isOpen || !useVcs) {
      setRepos([])
      setSelectedRepo('')
      setReposError(null)
      return
    }

    const fetchRepos = async () => {
      setLoadingRepos(true)
      setValidationError(null)
      setReposError(null)
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
          setReposError(null)
          // Auto-select first repo
          if (response.data.repositories.length > 0) {
            setSelectedRepo(response.data.repositories[0])
          }
        } else if (response.error) {
          setReposError(formatError(response.error))
          setRepos([])
          setSelectedRepo(null)
        } else {
          // Handle case where data exists but repositories is missing/invalid
          setRepos([])
          setSelectedRepo(null)
        }
      } catch (err) {
        setReposError(
          'Failed to load repositories. Please check your VCS connection.'
        )
        setRepos([])
        setSelectedRepo(null)
        console.error('Error loading repositories:', err)
      } finally {
        setLoadingRepos(false)
      }
    }

    fetchRepos()
  }, [selectedVcsConnectionId, isOpen, orgId, useVcs])

  // Fetch branches when repo changes
  useEffect(() => {
    if (!selectedRepo || !selectedVcsConnectionId || !isOpen || !useVcs) {
      setBranches([])
      setBranchesError(null)
      return
    }

    const [owner, repo] = selectedRepo.full_name.split('/')
    if (!owner || !repo) return

    const fetchBranches = async () => {
      setLoadingBranches(true)
      setValidationError(null)
      setBranchesError(null)
      try {
        const response = await getConnectionBranches(
          orgId,
          selectedVcsConnectionId,
          owner,
          repo
        )
        if (response.data) {
          setBranches(response.data)
          setBranchesError(null)
          // Auto-select 'main' if it exists, otherwise first branch
          const mainBranch = response.data.find((b) => b.name === 'main')
          if (mainBranch) {
            setSelectedBranch('main')
          } else if (response.data.length > 0) {
            setSelectedBranch(response.data[0].name)
          }
        } else if (response.error) {
          setBranchesError(formatError(response.error))
          setBranches([])
        }
      } catch (err) {
        setBranchesError('Failed to load branches. Please try again.')
        setBranches([])
        console.error('Error loading branches:', err)
      } finally {
        setLoadingBranches(false)
      }
    }

    fetchBranches()
  }, [selectedRepo, selectedVcsConnectionId, isOpen, orgId, useVcs])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
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
      connected_github_vcs_config?: any
      public_git_vcs_config?: any
    } = {
      name: name.trim(),
    }

    if (useVcs && selectedRepo) {
      // Use public_git_vcs_config for public repos, connected_github_vcs_config for private repos
      if (selectedRepo.private) {
        body.vcs_connection_id = selectedVcsConnectionId
        body.connected_github_vcs_config = {
          repo: selectedRepo.full_name,
          branch: selectedBranch,
          directory: directory.trim(),
        }

        // Only include path_filter if it's not empty
        if (pathFilter.trim()) {
          body.connected_github_vcs_config.path_filter = pathFilter.trim()
        }
      } else {
        // Public repo - use public_git_vcs_config
        body.public_git_vcs_config = {
          repo: selectedRepo.full_name,
          branch: selectedBranch,
          directory: directory.trim(),
        }

        // Only include path_filter if it's not empty
        if (pathFilter.trim()) {
          body.public_git_vcs_config.path_filter = pathFilter.trim()
        }
      }
    }

    await execute(orgId, appId, body)
  }

  const handleClose = () => {
    setName('')
    setUseVcs(true)
    setSelectedRepo(null)
    setSelectedBranch('main')
    setDirectory('.')
    setPathFilter('')
    setValidationError(null)
    setReposError(null)
    setBranchesError(null)
    onClose()
  }

  return (
    <ModalBase
      isVisible={isOpen}
      onClose={handleClose}
      heading="Create App Branch"
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        {validationError && <Banner theme="error">{validationError}</Banner>}

        <div className="flex flex-col gap-2">
          <label htmlFor="branch-name" className="text-sm font-medium">
            Branch Name
          </label>
          <Input
            id="branch-name"
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="production"
            required
            disabled={isLoading}
          />
        </div>

        <div className="flex items-center gap-2">
          <input
            id="use-vcs"
            type="checkbox"
            checked={useVcs}
            onChange={(e) => {
              setUseVcs(e.target.checked)
              // Clear VCS-related validation errors when toggled off
              if (!e.target.checked && validationError?.includes('VCS')) {
                setValidationError(null)
              }
            }}
            disabled={isLoading}
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
                No VCS connections found. Please connect your GitHub account
                first.
              </Banner>
            ) : (
              <>
                {vcsConnections.length > 1 && (
                  <div className="flex flex-col gap-2">
                    <label
                      htmlFor="vcs-connection"
                      className="text-sm font-medium"
                    >
                      VCS Connection
                    </label>
                    <select
                      id="vcs-connection"
                      value={selectedVcsConnectionId}
                      onChange={(e) =>
                        setSelectedVcsConnectionId(e.target.value)
                      }
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      disabled={isLoading || loadingRepos}
                    >
                      {vcsConnections.map((conn) => (
                        <option key={conn.id} value={conn.id}>
                          {conn.github_account_name ||
                            conn.github_install_id ||
                            conn.id}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {reposError && <Banner theme="error">{reposError}</Banner>}

                <div className="flex flex-col gap-2">
                  <label htmlFor="repo" className="text-sm font-medium">
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
                      id="repo"
                      value={selectedRepo?.full_name || ''}
                      onChange={(e) => {
                        const repo = repos.find(
                          (r) => r.full_name === e.target.value
                        )
                        setSelectedRepo(repo || null)
                      }}
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      required={useVcs}
                      disabled={
                        isLoading || loadingRepos || loadingBranches || !useVcs
                      }
                    >
                      {Array.isArray(repos) &&
                        repos.map((repo) => (
                          <option
                            key={repo.id || repo.full_name}
                            value={repo.full_name}
                          >
                            {repo.full_name}
                            {repo.private ? ' ��' : ''}
                          </option>
                        ))}
                    </select>
                  )}
                </div>

                {branchesError && (
                  <Banner theme="error">{branchesError}</Banner>
                )}

                <div className="flex flex-col gap-2">
                  <label htmlFor="git-branch" className="text-sm font-medium">
                    Git Branch
                  </label>
                  {loadingBranches ? (
                    <div className="px-3 py-2 text-sm text-gray-500">
                      Loading branches...
                    </div>
                  ) : branchesError ? (
                    <Input
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isLoading || !useVcs}
                    />
                  ) : branches.length === 0 && selectedRepo ? (
                    <div className="flex flex-col gap-2">
                      <div className="px-3 py-2 text-sm text-gray-500">
                        No branches found. Enter branch name manually.
                      </div>
                      <Input
                        id="git-branch"
                        type="text"
                        value={selectedBranch}
                        onChange={(e) => setSelectedBranch(e.target.value)}
                        placeholder="main"
                        required={useVcs}
                        disabled={isLoading || !useVcs}
                      />
                    </div>
                  ) : branches.length > 0 ? (
                    <select
                      id="git-branch"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      required={useVcs}
                      disabled={isLoading || loadingBranches || !useVcs}
                    >
                      {branches.map((branch) => (
                        <option key={branch.name} value={branch.name}>
                          {branch.name}
                        </option>
                      ))}
                    </select>
                  ) : (
                    <Input
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isLoading || !useVcs}
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
                    required
                    disabled={isLoading}
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
                    disabled={isLoading}
                  />
                  <p className="text-xs text-gray-500">
                    Regex pattern to filter which file changes trigger workflow
                    runs
                  </p>
                </div>
              </>
            )}
          </>
        )}

        <div className="flex gap-2 justify-end mt-4">
          <Button
            type="button"
            onClick={handleClose}
            variant="secondary"
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="primary"
            disabled={
              isLoading ||
              (useVcs &&
                (loadingRepos ||
                  loadingBranches ||
                  vcsConnections.length === 0 ||
                  !selectedRepo ||
                  !selectedBranch))
            }
          >
            {isLoading ? 'Creating...' : 'Create Branch'}
          </Button>
        </div>
      </form>
    </ModalBase>
  )
}
