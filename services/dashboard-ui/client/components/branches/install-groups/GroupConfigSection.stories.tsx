import { GroupConfigSection } from './GroupConfigSection'

export default { title: 'Branches/InstallGroups/GroupConfigSection' }

const mockInstalls = [
  { id: 'inst-1', name: 'Production US East' },
  { id: 'inst-2', name: 'Production EU West' },
  { id: 'inst-3', name: 'Staging' },
] as any[]

export const NoGroupSelected = () => (
  <div style={{ width: 320 }}>
    <GroupConfigSection
      group={undefined}
      availableInstalls={mockInstalls}
      onUpdate={() => {}}
      onDelete={() => {}}
    />
  </div>
)

export const WithGroup = () => (
  <div style={{ width: 320 }}>
    <GroupConfigSection
      group={{
        id: 'group-1',
        name: 'Canary',
        max_parallel: 2,
        requires_approval: true,
        rollback_on_failure: false,
        install_ids: ['inst-1', 'inst-2'],
      } as any}
      availableInstalls={mockInstalls}
      onUpdate={() => {}}
      onDelete={() => {}}
    />
  </div>
)
