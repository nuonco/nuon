export default {
  title: 'ServiceAccounts/CreateServiceAccount',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateServiceAccountModal } from './CreateServiceAccount'

const noop = () => {}

const roleOptions = [
  { value: 'org_admin', label: 'Admin' },
  { value: 'org_read_only', label: 'Read-only' },
  { value: 'runner', label: 'Runner' },
]

export const Default = () => (
  <ModalStory>
    <CreateServiceAccountModal
      roleOptions={roleOptions}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateServiceAccountModal
      roleOptions={roleOptions}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateServiceAccountModal
      roleOptions={roleOptions}
      isPending={false}
      error={{ error: 'role is required', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
