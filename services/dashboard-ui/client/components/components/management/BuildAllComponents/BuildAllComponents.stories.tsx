export default {
  title: 'Components/BuildAllComponents',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BuildAllComponentsModal } from './BuildAllComponents'

export const Default = () => (
  <ModalStory>
    <BuildAllComponentsModal
      appName="My App"
      isLoading={false}
      error={null}
      onBuild={() => {}}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <BuildAllComponentsModal
      appName="My App"
      isLoading={true}
      error={null}
      onBuild={() => {}}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <BuildAllComponentsModal
      appName="My App"
      isLoading={false}
      error={{ error: 'Unable to build components' } as any}
      onBuild={() => {}}
    />
  </ModalStory>
)
