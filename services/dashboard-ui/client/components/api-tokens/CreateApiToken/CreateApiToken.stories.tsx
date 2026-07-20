export default {
  title: 'ApiTokens/CreateApiToken',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateApiTokenModal } from './CreateApiToken'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateApiTokenModal
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
    <CreateApiTokenModal
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
    <CreateApiTokenModal
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
    <CreateApiTokenModal
      isPending={false}
      error={{ error: 'name is required', description: '', user_error: true }}
      createdToken={null}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)
