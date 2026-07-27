export default {
  title: 'Team/InviteUser',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TOrgInvite } from '@/types'
import { InviteUserModal } from './InviteUser'

const noop = () => {}

const pendingInvites = [
  {
    id: 'invxxxxxxxxxxxxxxxxxxxxxxxx',
    email: 'existing@email.com',
    role_type: 'org_admin',
    status: 'pending',
  },
] as TOrgInvite[]

export const Default = () => (
  <ModalStory>
    <InviteUserModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <InviteUserModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <InviteUserModal
      isPending={false}
      error={{ error: 'User already invited', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)

export const ExistingInvite = () => (
  <ModalStory>
    <InviteUserModal
      isPending={false}
      isResendPending={false}
      error={null}
      invites={pendingInvites}
      onSubmit={noop}
      onResend={noop}
    />
  </ModalStory>
)
