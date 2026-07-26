export default {
  title: 'ServiceAccounts/ChangeServiceAccountRole',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ChangeServiceAccountRoleModal } from './ChangeServiceAccountRole'

const noop = () => {}

const roleOptions = [
  { value: 'org_admin', label: 'Admin' },
  { value: 'org_read_only', label: 'Read-only' },
  { value: 'runner', label: 'Runner' },
]

export const Default = () => (
  <ModalStory>
    <ChangeServiceAccountRoleModal
      accountIdentity="svc-ci-deploy@example.com"
      currentRole="runner"
      roleOptions={roleOptions}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <ChangeServiceAccountRoleModal
      accountIdentity="svc-ci-deploy@example.com"
      currentRole="runner"
      roleOptions={roleOptions}
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ChangeServiceAccountRoleModal
      accountIdentity="svc-ci-deploy@example.com"
      currentRole="org_admin"
      roleOptions={roleOptions}
      isPending={false}
      error={{ error: 'Cannot change role', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
