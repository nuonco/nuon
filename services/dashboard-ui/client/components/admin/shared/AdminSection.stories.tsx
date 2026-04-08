export default {
  title: 'Admin/AdminSection',
}

import { AdminSection } from './AdminSection'
import { Button } from '@/components/common/Button'
import { Badge } from '@/components/common/Badge'

export const Default = () => (
  <AdminSection
    title="Runner management"
    subtitle="Manage and monitor runners for this organization"
  >
    <Button>Restart all runners</Button>
  </AdminSection>
)

export const WithMetadata = () => (
  <AdminSection
    title="Install details"
    subtitle="View and manage this install"
    metadata={
      <>
        <Badge theme="success">Active</Badge>
        <Badge theme="info">v2.3.1</Badge>
      </>
    }
  >
    <Button>Force redeploy</Button>
  </AdminSection>
)
