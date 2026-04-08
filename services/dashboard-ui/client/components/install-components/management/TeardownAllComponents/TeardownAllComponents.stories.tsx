export default {
  title: 'Install Components/TeardownAllComponents',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { TeardownAllComponentsModal } from './TeardownAllComponents'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <TeardownAllComponentsModal
      installName="production"
      isPending={false}
      isKickedOff={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <TeardownAllComponentsModal
      installName="production"
      isPending={true}
      isKickedOff={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <TeardownAllComponentsModal
      installName="production"
      isPending={false}
      isKickedOff={false}
      error={{ error: 'Unable to teardown components' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
