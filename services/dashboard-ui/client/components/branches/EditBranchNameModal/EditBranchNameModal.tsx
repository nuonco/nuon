import { useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
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
