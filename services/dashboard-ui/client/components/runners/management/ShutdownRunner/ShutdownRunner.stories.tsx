export default {
  title: 'Runners/ShutdownRunner',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ShutdownRunnerModal } from './ShutdownRunner'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <ShutdownRunnerModal
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithRunnerLabel = () => (
  <ModalStory>
    <ShutdownRunnerModal
      showRunnerLabel
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <ShutdownRunnerModal
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <ShutdownRunnerModal
      isPending={false}
      error={{ error: 'Unable to shutdown runner process.' } as any}
      onSubmit={noop}
      onClose={noop}
    />
  </ModalStory>
)
