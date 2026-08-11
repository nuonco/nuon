import { useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
import type {
  TAppBranch,
  TAppBranchConfig,
  TVCSConnectionRepo,
  TVCSBranch,
  TVCSConnection,
} from '@/types'
import { editBranchSchema, type EditBranchValues } from './schema'

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
  const [validationError, setValidationError] = useState<string | null>(null)

  const form = useForm({
    defaultValues: {
      branchName: branch.name || '',
      useVcs: !!(
        currentConfig?.connected_github_vcs_config ||
        currentConfig?.public_git_vcs_config
      ),
      directory:
        currentConfig?.connected_github_vcs_config?.directory ||
        currentConfig?.public_git_vcs_config?.directory ||
        '.',
      pathFilter:
        currentConfig?.connected_github_vcs_config?.path_filter ||
        currentConfig?.public_git_vcs_config?.path_filter ||
        '',
    } as EditBranchValues,
    validators: { onMount: editBranchSchema, onChange: editBranchSchema },
    onSubmit: ({ value }) => {
      setValidationError(null)

      if (value.useVcs && !selectedRepo) {
        setValidationError('Select a repository')
        return
      }

      onSubmit({
        branchName: value.branchName.trim(),
        useVcs: value.useVcs,
        selectedVcsConnectionId,
        selectedRepo,
        selectedBranch,
        directory: value.directory.trim(),
        pathFilter: value.pathFilter.trim(),
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const values = useStore(form.store, (s) => s.values)

  const displayError = externalValidationError || validationError

  return (
    <Modal
      heading="Edit branch"
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Saving changes' : 'Save changes',
        onClick: () => form.handleSubmit(),
        disabled: isSubmitting || !canSubmit,
        variant: 'primary',
      }}
      secondaryActionTrigger={{
        children: 'Cancel',
        onClick: onCancel,
        disabled: isSubmitting,
      }}
      {...props}
    >
      <form
        autoComplete="off"
        noValidate
        onSubmit={(e) => e.preventDefault()}
        className="flex flex-col gap-4"
      >
        {displayError && (
          <Banner theme="error" className="mb-4">
            {displayError}
          </Banner>
        )}

        <form.Field name="branchName">
          {(field) => (
            <FormInput
              field={field}
              id="branch-name"
              type="text"
              placeholder="Enter branch name"
              disabled={isSubmitting}
              autoFocus
              labelProps={{ labelText: 'Branch name' }}
            />
          )}
        </form.Field>

        <form.Field name="useVcs">
          {(field) => (
            <FormCheckbox
              field={field}
              id="use-vcs"
              disabled={isSubmitting}
              labelProps={{ labelText: 'Connect to git repository' }}
            />
          )}
        </form.Field>

        {values.useVcs && (
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
            directory={values.directory}
            onDirectoryChange={(v) => form.setFieldValue('directory', v)}
            pathFilter={values.pathFilter}
            onPathFilterChange={(v) => form.setFieldValue('pathFilter', v)}
            isSubmitting={isSubmitting}
          />
        )}
      </form>
    </Modal>
  )
}
