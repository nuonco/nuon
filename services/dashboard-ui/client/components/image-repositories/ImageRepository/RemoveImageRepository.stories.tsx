import { ModalStory } from '@/components/__stories__/helpers'
import { repositories } from '../fixtures'
import { RemoveImageRepository } from './RemoveImageRepository'

export default { title: 'Features / Image repositories / Remove' }

export const Confirm = () => (
  <ModalStory>
    <RemoveImageRepository
      name={repositories[0].image_url_prefix}
      isPending={false}
      onDelete={() => {}}
      isVisible
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <RemoveImageRepository
      name={repositories[1].image_url_prefix}
      isPending
      onDelete={() => {}}
      isVisible
    />
  </ModalStory>
)

export const Failed = () => (
  <ModalStory>
    <RemoveImageRepository
      name={repositories[2].image_url_prefix}
      isPending={false}
      onDelete={() => {}}
      error={{
        error: 'Conflict',
        description: 'The repository could not be removed. Try again.',
        user_error: true,
        status: 409,
      }}
      isVisible
    />
  </ModalStory>
)
