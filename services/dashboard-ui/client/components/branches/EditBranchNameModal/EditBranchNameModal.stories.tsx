export default {
  title: 'Branches/EditBranchNameModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EditBranchNameModal } from './EditBranchNameModal'

const noop = () => {}

const mockBranch = {
  id: 'br-001',
  name: 'production',
  created_at: '2024-01-15T10:30:00Z',
  workflows: [],
} as any

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
    <EditBranchNameModal
      branch={mockBranch}
      vcsConnections={[]}
      isSubmitting={false}
      validationError={null}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const WithExistingConfig = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      currentConfig={{
        connected_github_vcs_config: {
          vcs_connection_id: 'conn-1',
          repo: 'my-org/my-repo',
          branch: 'main',
          directory: '.',
        },
      } as any}
      vcsConnections={[
        {
          id: 'conn-1',
          github_account_name: 'my-org',
          github_install_id: '12345',
        } as any,
      ]}
      isSubmitting={false}
      validationError={null}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
      selectedVcsConnectionId="conn-1"
    />
  </ModalStory>
)

export const Submitting = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      vcsConnections={[]}
      isSubmitting={true}
      validationError={null}
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)

export const WithValidationError = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      vcsConnections={[]}
      isSubmitting={false}
      validationError="Branch name already exists"
      onSubmit={noop}
      onCancel={noop}
      {...defaultVcsProps}
    />
  </ModalStory>
)
