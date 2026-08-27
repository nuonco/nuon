import { useEffect, useRef, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
import type {
  TAPIError,
  TVCSConnectionRepo,
  TVCSBranch,
  TVCSConnection,
} from '@/types'
import {
  branchFormSchema,
  type BranchFormMode,
  type BranchFormOutput,
  type BranchFormValues,
} from './schema'

interface IBranchFormModal extends Omit<IModal, 'onSubmit'> {
  mode: BranchFormMode
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
  defaultName?: string
  defaultUseVcs?: boolean
  defaultDirectory?: string
  defaultPathFilter?: string
  defaultDisableBranchTriggers?: boolean
  isSubmitting: boolean
  submitError?: TAPIError | Error | null
  onSubmit: (output: BranchFormOutput) => void
  onCancel: () => void
}

export const BranchFormModal = ({
  mode,
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
  defaultName = '',
  defaultUseVcs = true,
  defaultDirectory = '.',
  defaultPathFilter = '',
  defaultDisableBranchTriggers = false,
  isSubmitting,
  submitError,
  onSubmit,
  onCancel,
  ...props
}: IBranchFormModal) => {
  const [repoError, setRepoError] = useState<string | null>(null)
  const bannerRef = useRef<HTMLDivElement>(null)

  const form = useForm({
    defaultValues: {
      name: defaultName,
      useVcs: defaultUseVcs,
      directory: defaultDirectory,
      pathFilter: defaultPathFilter,
      disableBranchTriggers: defaultDisableBranchTriggers,
    } as BranchFormValues,
    validators: { onMount: branchFormSchema, onChange: branchFormSchema },
    onSubmit: ({ value }) => {
      if (value.useVcs && !selectedRepo) {
        setRepoError(
          'Select a repository, or uncheck "Connect to git repository".'
        )
        return
      }
      setRepoError(null)

      onSubmit({
        name: value.name.trim(),
        useVcs: value.useVcs,
        selectedVcsConnectionId,
        selectedRepo,
        selectedBranch,
        directory: value.directory.trim(),
        pathFilter: value.pathFilter.trim(),
        disableBranchTriggers: value.disableBranchTriggers,
      })
    },
  })

  const canSubmit = useStore(form.store, (s) => s.canSubmit)
  const values = useStore(form.store, (s) => s.values)

  const submitErrorMessage =
    repoError ||
    (submitError
      ? ('error' in submitError && submitError.error) ||
        ('description' in submitError && submitError.description) ||
        ('message' in submitError && submitError.message) ||
        'Unable to save branch.'
      : undefined)

  useEffect(() => {
    if (submitErrorMessage && bannerRef.current) {
      bannerRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, [submitErrorMessage])

  return (
    <Modal
      heading={mode === 'create' ? 'Create app branch' : 'Edit branch'}
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? (
          <span className="flex items-center gap-2">
            <Icon variant="Loading" />
            {mode === 'create' ? 'Creating branch' : 'Saving changes'}
          </span>
        ) : mode === 'create' ? (
          <span className="flex items-center gap-2">
            <Icon variant="PlusIcon" />
            Create branch
          </span>
        ) : (
          'Save changes'
        ),
        disabled: !canSubmit || isSubmitting,
        onClick: () => form.handleSubmit(),
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
        {submitErrorMessage && (
          <div ref={bannerRef}>
            <Banner theme="error">{submitErrorMessage}</Banner>
          </div>
        )}

        <form.Field name="name">
          {(field) => (
            <FormInput
              field={field}
              id="branch-name"
              type="text"
              placeholder="production"
              disabled={isSubmitting}
              labelProps={{ labelText: 'Branch name' }}
            />
          )}
        </form.Field>

        <div className="flex flex-col gap-1">
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
          {!values.useVcs && (
            <Text variant="subtext" theme="neutral" className="px-2">
              Without a repository this branch cannot run. You can add one later
              from branch settings.
            </Text>
          )}
        </div>

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

        {mode === 'edit' ? (
          <div className="flex flex-col gap-1">
            <form.Field name="disableBranchTriggers">
              {(field) => (
                <FormCheckbox
                  field={field}
                  id="disable-branch-triggers"
                  disabled={isSubmitting}
                  labelProps={{ labelText: 'Disable webhook triggers' }}
                />
              )}
            </form.Field>
            <Text variant="subtext" theme="neutral" className="px-2">
              When enabled, git push and pull request events will not start runs
              on this branch.
            </Text>
          </div>
        ) : null}
      </form>
    </Modal>
  )
}
