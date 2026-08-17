export default {
  title: 'Notebooks/CreateNotebook',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { CreateNotebookModal } from './CreateNotebook'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <CreateNotebookModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <CreateNotebookModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateNotebookModal
      isPending={false}
      error={{
        error: 'A notebook with this name already exists.',
        description: '',
        user_error: true,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
