export default {
  title: 'Team/RemoveUser',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RemoveUserModal } from './RemoveUser'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RemoveUserModal
      accountEmail="alice@example.com"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RemoveUserModal
      accountEmail="alice@example.com"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RemoveUserModal
      accountEmail="alice@example.com"
      isPending={false}
      error={{ error: 'Cannot remove the last admin', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
