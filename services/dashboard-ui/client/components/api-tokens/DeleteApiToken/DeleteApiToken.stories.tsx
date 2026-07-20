export default {
  title: 'ApiTokens/DeleteApiToken',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteApiTokenModal } from './DeleteApiToken'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeleteApiTokenModal
      tokenName="ci-deploy"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteApiTokenModal
      tokenName="ci-deploy"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteApiTokenModal
      tokenName="ci-deploy"
      isPending={false}
      error={{ error: 'Token not found', description: '', user_error: true }}
      onSubmit={noop}
    />
  </ModalStory>
)
