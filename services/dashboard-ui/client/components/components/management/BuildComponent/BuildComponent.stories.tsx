export default {
  title: 'Components/BuildComponent',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { BuildComponentModal } from './BuildComponent'

const mockComponent = { id: 'comp-1', name: 'My Component' } as any

export const Default = () => (
  <ModalStory>
    <BuildComponentModal
      component={mockComponent}
      isLoading={false}
      error={null}
      onBuild={() => {}}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <BuildComponentModal
      component={mockComponent}
      isLoading={true}
      error={null}
      onBuild={() => {}}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <BuildComponentModal
      component={mockComponent}
      isLoading={false}
      error={{ error: 'Unable to build component' } as any}
      onBuild={() => {}}
    />
  </ModalStory>
)
