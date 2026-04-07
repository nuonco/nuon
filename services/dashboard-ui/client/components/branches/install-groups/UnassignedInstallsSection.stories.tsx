import type { Meta, StoryObj } from '@ladle/react'
import { UnassignedInstallsSection } from './UnassignedInstallsSection'

export default {
  title: 'Branches/InstallGroups/UnassignedInstallsSection',
} satisfies Meta

const mockInstalls = [
  { id: 'inst-1', name: 'Production US East' },
  { id: 'inst-2', name: 'Production EU West' },
  { id: 'inst-3', name: 'Staging' },
  { id: 'inst-4', name: 'Development' },
] as any[]

export const Default: StoryObj = {
  render: () => (
    <UnassignedInstallsSection
      installs={mockInstalls}
      assignedInstallIds={['inst-2']}
    />
  ),
}

export const AllAssigned: StoryObj = {
  render: () => (
    <UnassignedInstallsSection
      installs={mockInstalls}
      assignedInstallIds={['inst-1', 'inst-2', 'inst-3', 'inst-4']}
    />
  ),
}

export const Empty: StoryObj = {
  render: () => (
    <UnassignedInstallsSection
      installs={[]}
      assignedInstallIds={[]}
    />
  ),
}
