import { type FormEvent, useEffect, useRef, useState } from 'react'
import { Banner } from '@/components/common/Banner'
import { Input } from '@/components/common/form/Input'
import { CheckboxInput } from '@/components/common/form/CheckboxInput'
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
  initialDirectory,
  isSubmitting,
  submitError,
  onSubmit,
  onCancel,
  ...props
}: ICreateBranchModal) => {
  const [name, setName] = useState('')
  const [useVcs, setUseVcs] = useState(vcsConnections.length > 0)
  const [directory, setDirectory] = useState(initialDirectory || '.')
  const [pathFilter, setPathFilter] = useState('')
  const [repoError, setRepoError] = useState<string | null>(null)

  const formRef = useRef<HTMLFormElement>(null)
  const bannerRef = useRef<HTMLDivElement>(null)

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

  const handleFormSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()

    const form = e.currentTarget
    const firstInvalid = form.querySelector<HTMLElement>(
      ':invalid:not(fieldset):not(form)'
    )
    if (firstInvalid) {
      firstInvalid.scrollIntoView({ behavior: 'smooth', block: 'center' })
      firstInvalid.focus()
      form.reportValidity()
      return
    }

    if (useVcs && !selectedRepo) {
      setRepoError('Select a repository, or uncheck "Connect to git repository".')
      return
    }
    setRepoError(null)

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
        disabled: isSubmitting,
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
      <form
        ref={formRef}
        noValidate
        onSubmit={handleFormSubmit}
        className="flex flex-col gap-4"
      >
        {submitErrorMessage && (
          <div ref={bannerRef}>
            <Banner theme="error">{submitErrorMessage}</Banner>
          </div>
        )}

        <Input
          id="branch-name"
          name="name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="production"
          required
          disabled={isSubmitting}
          labelProps={{ labelText: 'Branch name' }}
        />

        <div className="flex flex-col gap-1">
          <CheckboxInput
            id="use-vcs"
            checked={useVcs}
            onChange={(e) => setUseVcs(e.target.checked)}
            disabled={isSubmitting}
            labelProps={{ labelText: 'Connect to git repository' }}
          />
          {!useVcs && (
            <Text variant="subtext" theme="neutral" className="px-2">
              Without a repository this branch cannot run. You can add one later
              from branch settings.
            </Text>
          )}
        </div>

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
      </form>
    </Modal>
  )
}
