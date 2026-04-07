export default {
  title: 'Install Components/DriftScanComponent',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DriftScanComponentModal } from './DriftScanComponent'
import type { TComponent } from '@/types'

const noop = () => {}

const mockComponent = {
  id: 'comp-abc123',
  name: 'web-server',
  type: 'helm_chart',
} as TComponent

const mockBuildSelect = () => <div className="p-4 border rounded">Build select placeholder</div>

export const Default = () => (
  <ModalStory>
    <DriftScanComponentModal
      component={mockComponent}
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <DriftScanComponentModal
      component={mockComponent}
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DriftScanComponentModal
      component={mockComponent}
      isPending={false}
      error={{ error: 'Drift scan failed: component not found' } as any}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
    />
  </ModalStory>
)
