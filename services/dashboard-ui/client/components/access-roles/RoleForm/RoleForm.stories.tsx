export default {
  title: 'Access roles/RoleForm',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { readAllWriteScoped } from '../permissions'
import { RoleFormModal } from './RoleForm'

const ORG_ID = 'orgrok933tcyzji01s7us3aeo3'
const noop = () => {}

export const Create = () => (
  <ModalStory>
    <RoleFormModal
      mode="create"
      orgId={ORG_ID}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Edit = () => (
  <ModalStory>
    <RoleFormModal
      mode="edit"
      orgId={ORG_ID}
      initialValues={{
        title: 'Release manager',
        description: 'Reads everything, deploys to production only.',
        contexts: ['team', 'api_token'],
        permissions: readAllWriteScoped({
          orgId: ORG_ID,
          installIds: ['inl4plkdhwau58atwfd92vlc8q'],
        }),
      }}
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RoleFormModal
      mode="create"
      orgId={ORG_ID}
      isPending
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RoleFormModal
      mode="create"
      orgId={ORG_ID}
      isPending={false}
      error={{
        error: 'a role named "Release manager" already exists in this org',
        description: '',
        user_error: true,
        status: 400,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
