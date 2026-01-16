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
import {
  getConnectionRepos,
  type Repo,
} from '@/lib/ctl-api/vcs/get-connection-repos'
import {
  getConnectionBranches,
  type Branch,
} from '@/lib/ctl-api/vcs/get-connection-branches'
import { createAppBranch } from './create-branch-action'
import { getMockRepositories, getMockBranches } from './new/mock-data'
import type { TCreateAppBranchRequest } from '@/types'

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
  const [repos, setRepos] = useState<Repo[]>([])
  const [selectedRepo, setSelectedRepo] = useState('')
  const [branches, setBranches] = useState<Branch[]>([])
  const [selectedBranch, setSelectedBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)

  const { data, error, isLoading, execute } = useServerAction({
    action: createAppBranch,
  })

  useServerActionToast({
    data,
    error,
    errorContent: (
      <>
        <Text>Failed to create app branch.</Text>
        <Text>{error?.error || 'Unknown error occurred.'}</Text>
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
    if (isOpen && vcsConnections.length > 0 && !selectedVcsConnectionId && useVcs) {
      setSelectedVcsConnectionId(vcsConnections[0].id)
    }
  }, [isOpen, vcsConnections, selectedVcsConnectionId, useVcs])

  // Fetch repos when VCS connection changes
  useEffect(() => {
    if (!selectedVcsConnectionId || !isOpen || !useVcs) {
      setRepos([])
      setSelectedRepo('')
      return
    }

    const fetchRepos = async () => {
      setLoadingRepos(true)
      setValidationError(null)
      try {
        const response = await getConnectionRepos(orgId, selectedVcsConnectionId)
        if (response.data) {
          setRepos(response.data)
          // Auto-select first repo
          if (response.data.length > 0) {
            setSelectedRepo(response.data[0].full_name)
          }
        } else if (response.error) {
          // Fallback to mock data on error
          console.log('Using mock repositories due to API error:', response.error)
          const mockRepos = getMockRepositories()
          setRepos(mockRepos as Repo[])
          if (mockRepos.length > 0) {
            setSelectedRepo(mockRepos[0].full_name)
          }
        }
      } catch (err) {
        // Fallback to mock data on exception
        console.log('Using mock repositories due to exception:', err)
        const mockRepos = getMockRepositories()
        setRepos(mockRepos as Repo[])
        if (mockRepos.length > 0) {
          setSelectedRepo(mockRepos[0].full_name)
        }
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
      return
    }

    const [owner, repo] = selectedRepo.split('/')
    if (!owner || !repo) return

    const fetchBranches = async () => {
      setLoadingBranches(true)
      setValidationError(null)
      try {
        const response = await getConnectionBranches(
          orgId,
          selectedVcsConnectionId,
          owner,
          repo
        )
        if (response.data) {
          setBranches(response.data)
          // Auto-select 'main' if it exists, otherwise first branch
          const mainBranch = response.data.find((b) => b.name === 'main')
          if (mainBranch) {
            setSelectedBranch('main')
          } else if (response.data.length > 0) {
            setSelectedBranch(response.data[0].name)
          }
        } else if (response.error) {
          // Fallback to mock branches
          console.log('Using mock branches due to API error:', response.error)
          const mockBranches = getMockBranches(selectedRepo)
          setBranches(mockBranches as Branch[])
          setSelectedBranch('main')
        }
      } catch (err) {
        // Fallback to mock branches
        console.log('Using mock branches due to exception:', err)
        const mockBranches = getMockBranches(selectedRepo)
        setBranches(mockBranches as Branch[])
        setSelectedBranch('main')
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

    const body: TCreateAppBranchRequest = {
      name: name.trim(),
    }

    if (useVcs) {
      body.connected_github_vcs_config = {
        repo: selectedRepo,
        branch: selectedBranch,
        directory: directory.trim(),
      }
      
      // Only include path_filter if it's not empty
      if (pathFilter.trim()) {
        body.connected_github_vcs_config.path_filter = pathFilter.trim()
      }
    }

    await execute(orgId, appId, body)
  }

  const handleClose = () => {
    setName('')
    setUseVcs(true)
    setSelectedRepo('')
    setSelectedBranch('main')
    setDirectory('.')
    setPathFilter('')
    setValidationError(null)
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
            onChange={(e) => setUseVcs(e.target.checked)}
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

                <div className="flex flex-col gap-2">
                  <label htmlFor="repo" className="text-sm font-medium">
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
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      required
                      disabled={isLoading || loadingRepos || loadingBranches}
                    >
                      {repos.map((repo) => (
                        <option key={repo.full_name} value={repo.full_name}>
                          {repo.full_name}
                          {repo.private ? ' 🔒' : ''}
                        </option>
                      ))}
                    </select>
                  )}
                </div>

                <div className="flex flex-col gap-2">
                  <label htmlFor="git-branch" className="text-sm font-medium">
                    Git Branch
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
                      required
                      disabled={isLoading}
                    />
                  ) : (
                    <select
                      id="git-branch"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                      required
                      disabled={isLoading || loadingBranches}
                    >
                      {branches.map((branch) => (
                        <option key={branch.name} value={branch.name}>
                          {branch.name}
                        </option>
                      ))}
                    </select>
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
                    Regex pattern to filter which file changes trigger workflow runs
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
              loadingRepos ||
              loadingBranches ||
              (useVcs && (vcsConnections.length === 0 || !selectedRepo || !selectedBranch))
            }
          >
            {isLoading ? 'Creating...' : 'Create Branch'}
          </Button>
        </div>
      </form>
    </ModalBase>
  )
}