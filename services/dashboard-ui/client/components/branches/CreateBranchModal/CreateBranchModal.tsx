import { useEffect, useRef, useState } from 'react'
import { useForm, useStore } from '@tanstack/react-form'
import { Banner } from '@/components/common/Banner'
import { FormCheckbox } from '@/components/common/form/FormCheckbox'
import { FormInput } from '@/components/common/form/FormInput'
import { Text } from '@/components/common/Text'
import { Modal, type IModal } from '@/components/surfaces/Modal'
import { BranchVcsConfigFields } from '@/components/branches/BranchVcsConfigFields'
import type {
  TAPIError,
  TCreateAppBranchRequest,
  TVCSConnectionRepo,
  TVCSBranch,
  TVCSConnection,
} from '@/types'
import { createBranchSchema, type CreateBranchValues } from './schema'

type TCreateBranchBody = TCreateAppBranchRequest & {
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
  initialDirectory?: string
  isSubmitting: boolean
  submitError?: TAPIError | Error | null
  onSubmit: (body: TCreateBranchBody) => void
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
  initialDirectory,
  isSubmitting,
  submitError,
  onSubmit,
  onCancel,
  ...props
}: ICreateBranchModal) => {
  const [repoError, setRepoError] = useState<string | null>(null)
  const bannerRef = useRef<HTMLDivElement>(null)

  const form = useForm({
    defaultValues: {
      name: '',
      useVcs: true,
      directory: initialDirectory || '.',
      pathFilter: '',
    } as CreateBranchValues,
    validators: { onMount: createBranchSchema, onChange: createBranchSchema },
    onSubmit: ({ value }) => {
      if (value.useVcs && !selectedRepo) {
        setRepoError(
          'Select a repository, or uncheck "Connect to git repository".'
        )
        return
      }
      setRepoError(null)

      const body: TCreateBranchBody = { name: value.name.trim() }

      if (value.useVcs && selectedRepo) {
        const config = {
          repo: selectedRepo.full_name,
          branch: selectedBranch,
          directory: value.directory.trim(),
          ...(value.pathFilter.trim()
            ? { path_filter: value.pathFilter.trim() }
            : {}),
        }
        if (selectedRepo.private) {
          body.vcs_connection_id = selectedVcsConnectionId
          body.connected_github_vcs_config = config
        } else {
          body.public_git_vcs_config = config
        }
      }

      onSubmit(body)
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
        'Unable to create branch.'
      : undefined)

  useEffect(() => {
    if (submitErrorMessage && bannerRef.current) {
      bannerRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' })
    }
  }, [submitErrorMessage])

  return (
    <Modal
      heading="Create app branch"
      size="lg"
      primaryActionTrigger={{
        children: isSubmitting ? 'Creating branch' : 'Create branch',
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
      </form>
    </Modal>
  )
}
