export default {
  title: 'Workflows/SkipStep',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { SkipStepModal } from './SkipStep'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <SkipStepModal isPending={false} error={null} onSubmit={noop} />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <SkipStepModal isPending={true} error={null} onSubmit={noop} />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <SkipStepModal
      isPending={false}
      error={{ error: 'Step cannot be skipped at this time' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
