export default {
  title: 'Team/InviteUser',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { InviteUserModal } from './InviteUser'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <InviteUserModal
      hasSupportRole={false}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithRoleSelection = () => (
  <ModalStory>
    <InviteUserModal
      hasSupportRole={true}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <InviteUserModal
      hasSupportRole={false}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <InviteUserModal
      hasSupportRole={false}
      isPending={false}
      error={{ error: 'User already invited', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
