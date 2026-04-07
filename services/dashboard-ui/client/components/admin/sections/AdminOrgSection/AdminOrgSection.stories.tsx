export default {
  title: 'Admin/AdminOrgSection',
}

import { AdminOrgSection } from './AdminOrgSection'

const mockOrg = {
  id: 'org-1',
  name: 'Test Org',
  features: {},
} as any

export const Default = () => (
  <AdminOrgSection
    orgId="org-1"
    org={mockOrg}
    adminEmail="admin@nuon.co"
    runner={{ id: 'runner-xyz789' } as any}
    runnerLoading={false}
  />
)

export const Loading = () => (
  <AdminOrgSection
    orgId="org-1"
    org={mockOrg}
    adminEmail="admin@nuon.co"
    runner={undefined}
    runnerLoading={true}
  />
)
