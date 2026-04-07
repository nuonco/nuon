export default {
  title: 'Branches/EditInstallGroupsModal',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { EditInstallGroupsModal } from './EditInstallGroupsModal'

const noop = () => {}

const mockInstalls = [
  { id: 'inst-1', name: 'production-us-east' },
  { id: 'inst-2', name: 'production-eu-west' },
  { id: 'inst-3', name: 'staging' },
  { id: 'inst-4', name: 'development' },
] as any[]

const mockGroups = [
  {
    id: 'group-1',
    name: 'Canary',
    install_ids: ['inst-3'],
    order: 0,
    max_parallel: 1,
    requires_approval: false,
    rollback_on_failure: true,
  },
  {
    id: 'group-2',
    name: 'Production',
    install_ids: ['inst-1', 'inst-2'],
    order: 1,
    max_parallel: 2,
    requires_approval: true,
    rollback_on_failure: true,
  },
]

export const Default = () => (
  <ModalStory>
    <EditInstallGroupsModal
      initialGroups={mockGroups}
      availableInstalls={mockInstalls}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <EditInstallGroupsModal
      initialGroups={[]}
      availableInstalls={mockInstalls}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const LoadingInstalls = () => (
  <ModalStory>
    <EditInstallGroupsModal
      initialGroups={[]}
      availableInstalls={[]}
      loadingInstalls={true}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const NoInstalls = () => (
  <ModalStory>
    <EditInstallGroupsModal
      initialGroups={[]}
      availableInstalls={[]}
      loadingInstalls={false}
      isSaving={false}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)

export const Saving = () => (
  <ModalStory>
    <EditInstallGroupsModal
      initialGroups={mockGroups}
      availableInstalls={mockInstalls}
      loadingInstalls={false}
      isSaving={true}
      onSave={noop}
      onCancel={noop}
    />
  </ModalStory>
)
