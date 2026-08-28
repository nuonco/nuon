export default {
  title: 'Apps/CreateApp',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateAppModal } from './CreateAppModal'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateAppModal isSubmitting={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateAppModal isSubmitting={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateAppModal
      isSubmitting={false}
      error={{
        error: 'An app with this name already exists.',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
