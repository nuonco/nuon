export default {
  title: 'Admin/AdminAppSection',
}

import { AdminAppSection } from './AdminAppSection'

export const Default = () => (
  <AdminAppSection
    orgId="org-1"
    appId="app-abc123"
    adminEmail="admin@nuon.co"
  />
)
