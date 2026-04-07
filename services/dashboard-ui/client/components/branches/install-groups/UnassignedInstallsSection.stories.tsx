import { UnassignedInstallsSection } from './UnassignedInstallsSection'

export default { title: 'Branches/InstallGroups/UnassignedInstallsSection' }

const mockInstalls = [
  { id: 'inst-1', name: 'Production US East' },
  { id: 'inst-2', name: 'Production EU West' },
  { id: 'inst-3', name: 'Staging' },
  { id: 'inst-4', name: 'Development' },
] as any[]

export const Default = () => (
  <UnassignedInstallsSection
    installs={mockInstalls}
    assignedInstallIds={['inst-2']}
  />
)

export const AllAssigned = () => (
  <UnassignedInstallsSection
    installs={mockInstalls}
    assignedInstallIds={['inst-1', 'inst-2', 'inst-3', 'inst-4']}
  />
)

export const Empty = () => (
  <UnassignedInstallsSection
    installs={[]}
    assignedInstallIds={[]}
  />
)
