export default {
  title: 'ApiTokens/CreateApiToken',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateApiTokenModal } from './CreateApiToken'

const noop = () => {}

const roleOptions = [
  { value: 'org_read_only', label: 'Read-only' },
  { value: 'org_admin', label: 'Admin' },
]

export const Default = () => (
  <ModalStory>
    <CreateApiTokenModal
      isPending={false}
      error={null}
      createdToken={null}
      roleOptions={roleOptions}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const RolesLoading = () => (
  <ModalStory>
    <CreateApiTokenModal
      isPending={false}
      error={null}
      createdToken={null}
      roleOptions={[]}
      rolesLoading={true}
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
      roleOptions={roleOptions}
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
      roleOptions={roleOptions}
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
      roleOptions={roleOptions}
      onSubmit={noop}
      onDone={noop}
    />
  </ModalStory>
)
