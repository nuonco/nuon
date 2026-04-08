export default {
  title: 'Install Components/TeardownComponent',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { TeardownComponentModal } from './TeardownComponent'
import type { TComponent } from '@/types'

const noop = () => {}

const mockComponent = {
  id: 'comp-abc123',
  name: 'web-server',
  type: 'helm_chart',
} as TComponent

const mockRoleSelector = () => <div className="p-4 border rounded">Role selector placeholder</div>

export const Default = () => (
  <ModalStory>
    <TeardownComponentModal
      component={mockComponent}
      isPending={false}
      error={null}
      onSubmit={noop}
      onClose={noop}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <TeardownComponentModal
      component={mockComponent}
      isPending={true}
      error={null}
      onSubmit={noop}
      onClose={noop}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <TeardownComponentModal
      component={mockComponent}
      isPending={false}
      error={{ error: 'Component teardown is already in progress' } as any}
      onSubmit={noop}
      onClose={noop}
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)
