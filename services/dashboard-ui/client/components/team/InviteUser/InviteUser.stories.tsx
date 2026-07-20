export default {
  title: 'Team/InviteUser',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { InviteUserModal } from './InviteUser'

const noop = () => {}

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
