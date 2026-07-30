export default { title: 'Triggers/Create trigger' }
import { ModalStory } from '@/components/__stories__/helpers'
import { CreateTriggerModal } from './CreateTrigger'
export const Default = () => (
  <ModalStory>
    <CreateTriggerModal
      error={null}
      isPending={false}
      onSubmit={() => undefined}
    />
  </ModalStory>
)
export const Failed = () => (
  <ModalStory>
    <CreateTriggerModal
      error={{
        description: 'Choose another name.',
        error: 'Name already exists',
        status: 409,
        user_error: true,
      }}
      isPending={false}
      onSubmit={() => undefined}
    />
  </ModalStory>
)
