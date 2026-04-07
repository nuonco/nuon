export default {
  title: 'Workflows/RetryStep',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { RetryStepModal } from './RetryStep'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <RetryStepModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <RetryStepModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <RetryStepModal
      isPending={false}
      error={{ error: 'Step cannot be retried' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
