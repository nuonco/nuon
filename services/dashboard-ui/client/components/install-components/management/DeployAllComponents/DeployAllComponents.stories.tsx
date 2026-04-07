export default {
  title: 'Install Components/DeployAllComponents',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeployAllComponentsModal } from './DeployAllComponents'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeployAllComponentsModal
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
    <DeployAllComponentsModal
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
    <DeployAllComponentsModal
      installName="production"
      isPending={false}
      isKickedOff={false}
      error={{ error: 'Unable to deploy components' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
