import { useNavigate } from 'react-router'
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
import { createAppBranch, getVCSConnectionRepos, getConnectionBranches } from '@/lib'
import type { TCreateAppBranchRequest, TVCSConnectionRepo, TVCSBranch } from '@/types'

type ICreateBranchModal = IModal

export const CreateBranchModal = ({ ...props }: ICreateBranchModal) => {
  const navigate = useNavigate()
  const { app } = useApp()
  const { org } = useOrg()
  const { addToast } = useToast()
  const { removeModal } = useSurfaces()

  const vcsConnections = org?.vcs_connections || []

  const [name, setName] = useState('')
  const [useVcs, setUseVcs] = useState(true)
  const [selectedVcsConnectionId, setSelectedVcsConnectionId] = useState(
    vcsConnections[0]?.id || ''
  )
  const [repos, setRepos] = useState<TVCSConnectionRepo[]>([])
  const [selectedRepo, setSelectedRepo] = useState<TVCSConnectionRepo | null>(
    null
  )
  const [branches, setBranches] = useState<TVCSBranch[]>([])
  const [selectedBranch, setSelectedBranch] = useState('main')
  const [directory, setDirectory] = useState('.')
  const [pathFilter, setPathFilter] = useState('')
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [loadingBranches, setLoadingBranches] = useState(false)
  const [validationError, setValidationError] = useState<string | null>(null)
  const [reposError, setReposError] = useState<string | null>(null)
  const [branchesError, setBranchesError] = useState<string | null>(null)

  const { mutate, isPending: isLoading } = useMutation({
    mutationFn: async (
      body: TCreateAppBranchRequest & {
        vcs_connection_id?: string
        connected_github_vcs_config?: any
        public_git_vcs_config?: any
      }
    ) => {
      return createAppBranch({ appId: app.id, body, orgId: org.id })
    },
    onSuccess: (data) => {
      addToast(
        <Toast heading="Branch created successfully" theme="success">
          <Text>Created app branch: {name}</Text>
        </Toast>
      )
      removeModal(props.modalId)
      navigate(`/${org.id}/apps/${app.id}/branches/${data.id}`)
    },
    onError: (error: Error) => {
      addToast(
        <Toast heading="Branch creation failed" theme="error">
          <Text>Failed to create app branch.</Text>
          <Text>{error.message || 'Unknown error occurred.'}</Text>
        </Toast>
      )
    },
  })

  useEffect(() => {
    if (!selectedVcsConnectionId || !useVcs) {
      setRepos([])
      setSelectedRepo(null)
      setReposError(null)
      return
    }

    setSelectedRepo(null)

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
            const sortedRepos = [...response.repositories].sort((a, b) =>
              a.full_name.localeCompare(b.full_name)
            )
            setRepos(sortedRepos)
            setReposError(null)
            if (sortedRepos.length > 0) {
              setSelectedRepo(sortedRepos[0])
            }
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

      setReposError(
        'Failed to load repositories. Please check your VCS connection.'
      )
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

    const [owner, repo] = selectedRepo.full_name.split('/')
    if (!owner || !repo) return

    const fetchBranches = async () => {
      setLoadingBranches(true)
      setValidationError(null)
      setBranchesError(null)

      let lastErr: unknown
      for (let attempt = 0; attempt < 3; attempt++) {
        if (attempt > 0) await new Promise((r) => setTimeout(r, 1000 * attempt))
        try {
          const fetchedBranches = await getConnectionBranches(
            org.id,
            selectedVcsConnectionId,
            owner,
            repo
          )
          setBranches(fetchedBranches)
          setBranchesError(null)
          const mainBranch = fetchedBranches.find((b) => b.name === 'main')
          if (mainBranch) {
            setSelectedBranch('main')
          } else if (fetchedBranches.length > 0) {
            setSelectedBranch(fetchedBranches[0].name)
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
      connected_github_vcs_config?: any
      public_git_vcs_config?: any
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

    mutate(body)
  }

  return (
    <Modal
      heading="Create App Branch"
      primaryActionTrigger={{
        children: isLoading ? 'Creating...' : 'Create branch',
        disabled:
          isLoading ||
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
        onClick: () => removeModal(props.modalId),
        disabled: isLoading,
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
          disabled={isLoading}
          labelProps={{ labelText: 'Branch Name' }}
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
          disabled={isLoading}
          labelProps={{ labelText: 'Connect to Git Repository' }}
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
                    onChange={(e) => setSelectedVcsConnectionId(e.target.value)}
                    disabled={isLoading || loadingRepos}
                    options={vcsConnections.map((conn) => ({
                      value: conn.id,
                      label:
                        conn.github_account_name ||
                        conn.github_install_id ||
                        conn.id,
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
                      setSelectedRepo(repo || null)
                    }}
                    required={useVcs}
                    disabled={
                      isLoading || loadingRepos || loadingBranches || !useVcs
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
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isLoading || !useVcs}
                      labelProps={{ labelText: 'Git Branch' }}
                    />
                  ) : branches.length === 0 && selectedRepo ? (
                    <Input
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isLoading || !useVcs}
                      labelProps={{ labelText: 'Git Branch' }}
                      helperText="No branches found. Enter branch name manually."
                    />
                  ) : branches.length > 0 ? (
                    <Select
                      id="git-branch"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      required={useVcs}
                      disabled={isLoading || loadingBranches || !useVcs}
                      options={branches.map((b) => ({
                        value: b.name,
                        label: b.name,
                      }))}
                      labelProps={{ labelText: 'Git Branch' }}
                      searchable
                    />
                  ) : (
                    <Input
                      id="git-branch"
                      type="text"
                      value={selectedBranch}
                      onChange={(e) => setSelectedBranch(e.target.value)}
                      placeholder="main"
                      required={useVcs}
                      disabled={isLoading || !useVcs}
                      labelProps={{ labelText: 'Git Branch' }}
                    />
                  ))}

                <Input
                  id="directory"
                  type="text"
                  value={directory}
                  onChange={(e) => setDirectory(e.target.value)}
                  placeholder="."
                  required
                  disabled={isLoading}
                  labelProps={{ labelText: 'Directory' }}
                  helperText='Path to your application config (use "." for root)'
                />

                <Input
                  id="path-filter"
                  type="text"
                  value={pathFilter}
                  onChange={(e) => setPathFilter(e.target.value)}
                  placeholder="^(src/|config/).*"
                  disabled={isLoading}
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

export const CreateBranchButton = ({
  ...props
}: Omit<IButtonAsButton, 'children'>) => {
  const { addModal } = useSurfaces()
  const modal = <CreateBranchModal />
  return (
    <Button variant="secondary" onClick={() => addModal(modal)} {...props}>
      <Icon variant="Plus" size={16} />
      Create branch
    </Button>
  )
}
