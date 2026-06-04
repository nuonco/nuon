export default {
  title: 'Install Components/DeployAllComponents',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeployAllComponentsModal } from './DeployAllComponents'

const noop = () => {}

const mockRoleSelector = () => <div className="p-4 border rounded">Role selector placeholder</div>

export const Default = () => (
  <ModalStory>
    <DeployAllComponentsModal
      installName="production"
      isPending={false}
      isKickedOff={false}
      error={null}
      onSubmit={noop}
      roleSelector={mockRoleSelector}
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
      roleSelector={mockRoleSelector}
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
      roleSelector={mockRoleSelector}
    />
  </ModalStory>
)
