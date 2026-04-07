export default {
  title: 'Branches/CreateBranchModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateBranchModal } from './CreateBranchModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateBranchModal
      orgId="org-1"
      vcsConnections={[]}
      isSubmitting={false}
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const WithVCSConnection = () => (
  <ModalStory>
    <CreateBranchModal
      orgId="org-1"
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
    />
  </ModalStory>
)

export const MultipleVCSConnections = () => (
  <ModalStory>
    <CreateBranchModal
      orgId="org-1"
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
    />
  </ModalStory>
)

export const Submitting = () => (
  <ModalStory>
    <CreateBranchModal
      orgId="org-1"
      vcsConnections={[]}
      isSubmitting={true}
      onSubmit={noop}
      onCancel={noop}
    />
  </ModalStory>
)
