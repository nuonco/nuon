import { AdminDashboardLink } from './AdminDashboardLink'

export default {
  title: 'Admin/AdminDashboardLink',
}

export const Default = () => (
  <AdminDashboardLink
    path="/queues?owner_id=inst123&owner_type=installs"
    label="View queues"
    isVisible={true}
    adminDashboardUrl="http://localhost:8085"
  />
)

export const Hidden = () => (
  <AdminDashboardLink
    path="/queues"
    label="View queues"
    isVisible={false}
    adminDashboardUrl="http://localhost:8085"
  />
)
