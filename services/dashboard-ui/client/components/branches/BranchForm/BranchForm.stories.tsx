export default {
  title: 'Branches/BranchForm',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BranchFormModal } from './BranchForm'

const noop = () => {}

const defaultVcsProps = {
  repos: [],
  branches: [],
  loadingRepos: false,
  loadingBranches: false,
  reposError: null,
  branchesError: null,
  selectedVcsConnectionId: '',
  onVcsConnectionChange: noop,
  selectedRepo: null,
  onRepoChange: noop,
  selectedBranch: 'main',
  onBranchChange: noop,
}

export const Create = () => (
  <ModalStory>
    <BranchFormModal
      mode="create"
      vcsConnections={[]}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const CreateWithVCSConnection = () => (
  <ModalStory>
    <BranchFormModal
      mode="create"
      vcsConnections={[
        {
          id: 'conn-1',
          github_account_name: 'my-org',
          github_install_id: '12345',
        } as any,
      ]}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
      selectedVcsConnectionId="conn-1"
    />
  </ModalStory>
)

export const CreateSubmitting = () => (
  <ModalStory>
    <BranchFormModal
      mode="create"
      vcsConnections={[]}
      isSubmitting={true}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const CreateSubmitError = () => (
  <ModalStory>
    <BranchFormModal
      mode="create"
      vcsConnections={[]}
      isSubmitting={false}
      submitError={{ error: 'A branch with this name already exists.' } as any}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const Edit = () => (
  <ModalStory>
    <BranchFormModal
      mode="edit"
      vcsConnections={[]}
      defaultName="production"
      defaultUseVcs={false}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const EditWithExistingConfig = () => (
  <ModalStory>
    <BranchFormModal
      mode="edit"
      vcsConnections={[
        {
          id: 'conn-1',
          github_account_name: 'my-org',
          github_install_id: '12345',
        } as any,
      ]}
      defaultName="production"
      defaultUseVcs={true}
      defaultDirectory="."
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
      selectedVcsConnectionId="conn-1"
    />
  </ModalStory>
)

export const EditSubmitError = () => (
  <ModalStory>
    <BranchFormModal
      mode="edit"
      vcsConnections={[]}
      defaultName="production"
      defaultUseVcs={false}
      isSubmitting={false}
      submitError={{ error: 'Branch name already exists' } as any}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)
