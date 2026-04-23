export default {
  title: 'Actions/InstallActionsTable',
}

import { Button } from '@/components/common/Button'
import { InstallActionsTable, InstallActionsTableSkeleton, type InstallActionRow } from './InstallActionsTable'

const mockRows: InstallActionRow[] = Array.from({ length: 3 }, (_, i) => ({
  actionId: `action-${i + 1}`,
  actionName: `deploy-step-${i + 1}`,
  actionStatus: <span className="text-sm">active</span>,
  actionTrigger: <span className="text-sm">post-deploy-component</span>,
  actionRunDatetime: <span className="text-sm">2 hours ago</span>,
  actionRunDuration: <span className="text-sm">1m 30s</span>,
  labels: i === 0 ? <span className="text-xs font-mono">category: setup</span> : null,
  href: `/org-1/installs/install-1/actions/action-${i + 1}`,
}))

export const Default = () => (
  <InstallActionsTable
    data={mockRows}
    filterActions={<Button variant="ghost">Run action</Button>}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <InstallActionsTable
    data={[]}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <InstallActionsTableSkeleton />
