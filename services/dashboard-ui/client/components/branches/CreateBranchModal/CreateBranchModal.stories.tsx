export default {
  title: 'Branches/CreateBranchModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateBranchModal } from './CreateBranchModal'

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

export const Default = () => (
  <ModalStory>
    <CreateBranchModal
      vcsConnections={[]}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const WithVCSConnection = () => (
  <ModalStory>
    <CreateBranchModal
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

export const MultipleVCSConnections = () => (
  <ModalStory>
    <CreateBranchModal
      vcsConnections={[
        {
          id: 'conn-1',
          github_account_name: 'my-org',
          github_install_id: '12345',
        } as any,
        {
          id: 'conn-2',
          github_account_name: 'other-org',
          github_install_id: '67890',
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

export const Submitting = () => (
  <ModalStory>
    <CreateBranchModal
      vcsConnections={[]}
      isSubmitting={true}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)
