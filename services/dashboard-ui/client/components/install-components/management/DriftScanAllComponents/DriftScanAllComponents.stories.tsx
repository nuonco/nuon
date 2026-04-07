export default {
  title: 'Install Components/DriftScanAllComponents',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DriftScanAllComponentsModal } from './DriftScanAllComponents'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DriftScanAllComponentsModal
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
    <DriftScanAllComponentsModal
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
    <DriftScanAllComponentsModal
      installName="production"
      isPending={false}
      isKickedOff={false}
      error={{ error: 'Unable to deploy components' } as any}
      onSubmit={noop}
    />
  </ModalStory>
)
