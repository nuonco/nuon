export default {
  title: 'Team/ChangeRole',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ChangeRoleModal } from './ChangeRole'

const noop = () => {}

const roleOptions = [
  { value: 'org_admin', label: 'Admin' },
  { value: 'org_read_only', label: 'Read-only' },
]

export const Default = () => (
  <ModalStory>
    <ChangeRoleModal
      accountEmail="user@example.com"
      currentRole="org_admin"
      roleOptions={roleOptions}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const ExistingSupportMember = () => (
  <ModalStory>
    <ChangeRoleModal
      accountEmail="user@example.com"
      currentRole="org_support"
      currentRoleLabel="Support"
      roleOptions={roleOptions}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <ChangeRoleModal
      accountEmail="user@example.com"
      currentRole="org_read_only"
      roleOptions={roleOptions}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ChangeRoleModal
      accountEmail="user@example.com"
      currentRole="org_admin"
      roleOptions={roleOptions}
      isPending={false}
      error={{
        error: 'Cannot demote the last org admin',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
