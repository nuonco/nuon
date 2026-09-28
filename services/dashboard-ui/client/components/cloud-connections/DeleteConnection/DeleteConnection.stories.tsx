import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteConnection } from './DeleteConnection'
export default { title: 'Cloud connections/Delete' }
export const Confirm = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-production"
      isPending={false}
      onDelete={() => {}}
    />
  </ModalStory>
)
export const Conflict = () => (
  <ModalStory>
    <DeleteConnection
      name="acme-production"
      isPending={false}
      onDelete={() => {}}
      error={{
        error: 'Conflict',
        description:
          'Cloud connection cannot be deleted; it is referenced by installs (2)',
        user_error: true,
        status: 409,
      }}
    />
  </ModalStory>
)
