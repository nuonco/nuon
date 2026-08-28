import { type FormEvent, useRef, useState } from 'react'
import { Badge } from '@/components/common/Badge'
import { Banner } from '@/components/common/Banner'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
import type {
  TAPIError,
  TAppInstallsConfig,
  TVCSBranch,
  TVCSConnection,
  TVCSConnectionRepo,
} from '@/types'

interface IManageInstallsConfig extends Omit<IModal, 'onSubmit'> {
  currentConfig?: TAppInstallsConfig
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
  submitError?: TAPIError | Error | null
  onSubmit: (body: {
    vcs_type: 'connected' | 'public'
    vcs_connection_id?: string
    repo: string
    branch: string
    directory: string
  }) => void
  onCancel: () => void
}

export const ManageInstallsConfig = ({
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
  submitError,
  onSubmit,
  onCancel,
  ...props
}: IManageInstallsConfig) => {
  const [directory, setDirectory] = useState(currentConfig?.directory || '.')
  const formRef = useRef<HTMLFormElement>(null)

  const submitErrorMessage = submitError
    ? ('error' in submitError && submitError.error) ||
      ('message' in submitError && submitError.message) ||
      'Unable to save installs config.'
    : undefined

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    if (!selectedRepo || !selectedBranch) return

    onSubmit({
      vcs_type: selectedRepo.private ? 'connected' : 'public',
      vcs_connection_id: selectedRepo.private
        ? selectedVcsConnectionId
        : undefined,
      repo: selectedRepo.full_name,
      branch: selectedBranch,
      directory: directory.trim() || '.',
    })
  }

  return (
    <Modal
      heading={
        currentConfig ? 'Update installs config' : 'Configure installs config'
      }
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving...' : 'Save',
        disabled: isSubmitting || !selectedRepo || !selectedBranch,
        onClick: () => formRef.current?.requestSubmit(),
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: onCancel,
        disabled: isSubmitting,
      }}
      {...props}
    >
      {currentConfig && (
        <div className="flex items-center gap-2 mb-4">
          <Text variant="subtext" theme="neutral">
            Current source:
          </Text>
          <Badge variant="code" size="md">
            {currentConfig.source === 'config' ? 'installs.toml' : 'dashboard'}
          </Badge>
          <Text variant="subtext" theme="neutral">
            {currentConfig.repo} / {currentConfig.branch}
          </Text>
        </div>
      )}

      <form
        ref={formRef}
        noValidate
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        {submitErrorMessage && (
          <Banner theme="error">{submitErrorMessage}</Banner>
        )}

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
          isSubmitting={isSubmitting}
        />
      </form>
    </Modal>
  )
}
