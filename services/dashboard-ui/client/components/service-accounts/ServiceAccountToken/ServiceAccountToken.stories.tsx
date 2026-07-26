export default {
  title: 'ServiceAccounts/ServiceAccountToken',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateServiceAccountTokenModal } from './ServiceAccountToken'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateServiceAccountTokenModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={false}
      error={null}
      createdToken={null}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateServiceAccountTokenModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={true}
      error={null}
      createdToken={null}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const Created = () => (
  <ModalStory>
    <CreateServiceAccountTokenModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={false}
      error={null}
      createdToken="nuon_tok_abc123def456ghi789"
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateServiceAccountTokenModal
      accountIdentity="svc-ci-deploy@example.com"
      isPending={false}
      error={{ error: 'duration is invalid', description: '', user_error: true }}
      createdToken={null}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)
