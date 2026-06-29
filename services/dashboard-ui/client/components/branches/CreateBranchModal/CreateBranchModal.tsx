import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
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
      size="lg"
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
          <BranchVcsConfigFields
            vcsConnections={vcsConnections}
            repos={repos}
            branches={branches}
            loadingRepos={loadingRepos}
            loadingBranches={loadingBranches}
            reposError={reposError}
            branchesError={branchesError}
            selectedVcsConnectionId={selectedVcsConnectionId}
            onVcsConnectionChange={onVcsConnectionChange}
            selectedRepo={selectedRepo}
            onRepoChange={onRepoChange}
            selectedBranch={selectedBranch}
            onBranchChange={onBranchChange}
            directory={directory}
            onDirectoryChange={setDirectory}
            pathFilter={pathFilter}
            onPathFilterChange={setPathFilter}
            isSubmitting={isSubmitting}
          />
        )}
      </div>
    </Modal>
  )
}
