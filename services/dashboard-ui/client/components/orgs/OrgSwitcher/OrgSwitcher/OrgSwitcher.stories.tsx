export default {
  title: 'Orgs/OrgSwitcher',
}

import { OrgSwitcher } from './OrgSwitcher'

const mockOrg = {
  id: 'org-1',
  name: 'My Organization',
  vcs_connections: [],
} as any

export const Default = () => (
  <div className="w-[248px] p-4">
    <OrgSwitcher org={mockOrg} isSidebarOpen />
  </div>
)

export const Collapsed = () => (
  <div className="w-[60px] p-4">
    <OrgSwitcher org={mockOrg} isSidebarOpen={false} />
  </div>
)
