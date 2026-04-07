export default {
  title: 'Admin/AdminActionGroup',
}

import { AdminActionGroup } from './AdminActionGroup'
import { Button } from '@/components/common/Button'

export const Default = () => (
  <AdminActionGroup title="Runner management" icon="Server">
    <Button variant="ghost">Restart runner</Button>
    <Button variant="ghost">View logs</Button>
  </AdminActionGroup>
)

export const Warning = () => (
  <AdminActionGroup title="Maintenance actions" icon="Warning" variant="warning">
    <Button variant="ghost">Disable auto-deploys</Button>
    <Button variant="ghost">Pause queue</Button>
  </AdminActionGroup>
)

export const Danger = () => (
  <AdminActionGroup title="Destructive actions" icon="Trash" variant="danger">
    <Button variant="danger">Delete all data</Button>
    <Button variant="danger">Force terminate</Button>
  </AdminActionGroup>
)
