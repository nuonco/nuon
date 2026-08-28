export default {
  title: 'ServiceAccounts/DeleteServiceAccount',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteServiceAccountModal } from './DeleteServiceAccount'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeleteServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={false}
      error={{
        error: 'Cannot delete the last admin service account',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
