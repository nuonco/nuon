export default {
  title: 'Runners/UpdateRunner',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { UpdateRunnerModal } from './UpdateRunner'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <UpdateRunnerModal
      isPending={false}
      error={{ error: 'Unable to update runner.' } as any}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
