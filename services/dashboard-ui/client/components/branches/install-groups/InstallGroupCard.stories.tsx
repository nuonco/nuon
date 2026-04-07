import { InstallGroupCard } from './InstallGroupCard'

export default { title: 'Branches/InstallGroups/InstallGroupCard' }

const mockInstalls = [
  { id: 'inst-1', name: 'Production US East' },
  { id: 'inst-2', name: 'Production EU West' },
] as any[]

export const Default = () => (
  <InstallGroupCard
    group={{
      id: 'group-1',
      name: 'Canary',
      max_parallel: 1,
      requires_approval: false,
      rollback_on_failure: false,
      install_ids: ['inst-1', 'inst-2'],
    } as any}
    installs={mockInstalls}
    isSelected={false}
    onClick={() => {}}
    onRemoveInstall={() => {}}
    index={0}
  />
)

export const Selected = () => (
  <InstallGroupCard
    group={{
      id: 'group-2',
      name: 'Production',
      max_parallel: 3,
      requires_approval: true,
      rollback_on_failure: true,
      install_ids: ['inst-1', 'inst-2'],
    } as any}
    installs={mockInstalls}
    isSelected
    onClick={() => {}}
    onRemoveInstall={() => {}}
    index={1}
  />
)

export const Empty = () => (
  <InstallGroupCard
    group={{
      id: 'group-3',
      name: 'New group',
      max_parallel: 1,
      requires_approval: false,
      rollback_on_failure: false,
      install_ids: [],
    } as any}
    installs={[]}
    isSelected={false}
    onClick={() => {}}
    onRemoveInstall={() => {}}
    index={2}
  />
)
