import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteRoleModal } from './DeleteRole'

export default {
  title: 'Access roles/DeleteRole',
}

export const Default = () => (
  <ModalStory>
    <DeleteRoleModal
      roleTitle="Release manager"
      isPending={false}
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const Deleting = () => (
  <ModalStory>
    <DeleteRoleModal
      roleTitle="Release manager"
      isPending
      error={null}
      onSubmit={() => {}}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteRoleModal
      roleTitle="Release manager"
      isPending={false}
      error={{ error: 'managed roles cannot be edited or deleted' } as any}
      onSubmit={() => {}}
    />
  </ModalStory>
)
