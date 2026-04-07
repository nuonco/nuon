export default {
  title: 'Install Components/DeployComponent',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeployComponentModal } from './DeployComponent'
import type { TComponent } from '@/types'

const noop = () => {}

const mockComponent = {
  id: 'comp-abc123',
  name: 'web-server',
  type: 'helm_chart',
} as TComponent

const mockBuildSelect = () => <div className="p-4 border rounded">Build select placeholder</div>
const mockRoleSelector = () => <div className="p-4 border rounded">Role selector placeholder</div>

export const Default = () => (
  <ModalStory>
    <DeployComponentModal
      component={mockComponent}
      installId="inst-abc123"
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <DeployComponentModal
      component={mockComponent}
      installId="inst-abc123"
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeployComponentModal
      component={mockComponent}
      installId="inst-abc123"
      isPending={false}
      error={{ error: 'Build not found or is no longer active' } as any}
      onSubmit={noop}
      onClose={noop}
      buildSelect={mockBuildSelect}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)
