export default {
  title: 'ServiceAccounts/RenameServiceAccount',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RenameServiceAccountModal } from './RenameServiceAccount'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RenameServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      currentName="ci-deploy"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RenameServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      currentName="ci-deploy"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RenameServiceAccountModal
      accountIdentity="svc-ci-deploy@example.com"
      currentName="ci-deploy"
      isPending={false}
      error={{ error: 'name is required', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
