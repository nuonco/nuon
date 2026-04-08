export default {
  title: 'Orgs/OrgSummary',
}

import { SidebarProvider } from '@/providers/sidebar-provider'
import { OrgSummary } from './OrgSummary'
import type { TOrg } from '@/types'

const mockOrg: TOrg = {
  id: 'org-1',
  name: 'Acme Corp',
  status: 'active',
  sandbox_mode: false,
} as TOrg

const sandboxOrg: TOrg = {
  ...mockOrg,
  id: 'org-2',
  name: 'Dev Sandbox',
  sandbox_mode: true,
}

export const Default = () => (
  <SidebarProvider>
    <div className="w-64 p-4">
      <OrgSummary org={mockOrg} isSidebarOpen />
    </div>
  </SidebarProvider>
)

export const Collapsed = () => (
  <SidebarProvider>
    <div className="w-16 p-4">
      <OrgSummary org={mockOrg} isSidebarOpen={false} isButtonSummary />
    </div>
  </SidebarProvider>
)

export const SandboxMode = () => (
  <SidebarProvider>
    <div className="w-64 p-4">
      <OrgSummary org={sandboxOrg} isSidebarOpen />
    </div>
  </SidebarProvider>
)
