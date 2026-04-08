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

export const Default = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      orgId="org-1"
      vcsConnections={[]}
      isSubmitting={false}
      validationError={null}
      onSubmit={noop}
      onCancel={noop}
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
      orgId="org-1"
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
    />
  </ModalStory>
)

export const Submitting = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      orgId="org-1"
      vcsConnections={[]}
      isSubmitting={true}
      validationError={null}
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const WithValidationError = () => (
  <ModalStory>
    <EditBranchNameModal
      branch={mockBranch}
      orgId="org-1"
      vcsConnections={[]}
      isSubmitting={false}
      validationError="Branch name already exists"
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)
