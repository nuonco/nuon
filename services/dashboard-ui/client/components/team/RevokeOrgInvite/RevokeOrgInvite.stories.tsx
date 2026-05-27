export default {
  title: 'Team/RevokeOrgInvite',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RevokeOrgInviteModal } from './RevokeOrgInvite'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RevokeOrgInviteModal
      email="pending@example.com"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RevokeOrgInviteModal
      email="pending@example.com"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RevokeOrgInviteModal
      email="pending@example.com"
      isPending={false}
      error={{ error: 'Only org admins can revoke invites', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
