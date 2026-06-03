export default {
  title: 'Actions/ActionsTable',
}

import { LabelBadge } from '@/components/common/LabelBadge'
import { ActionsTable, ActionsTableSkeleton, type TActionRow } from './ActionsTable'

const mockRows: TActionRow[] = Array.from({ length: 3 }, (_, i) => ({
  actionId: `action-${i + 1}`,
  actionName: `deploy-step-${i + 1}`,
  labels: i === 0 ? <LabelBadge labelKey="category" labelValue="setup" size="sm" /> : null,
  actionTriggers: <span className="text-sm">post-deploy-component</span>,
  actionSteps: (
    <ol className="flex flex-col gap-1 list-decimal">
      <li className="text-sm">run script</li>
    </ol>
  ),
  href: `/org-1/apps/app-1/actions/action-${i + 1}`,
}))

export const Default = () => (
  <ActionsTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <ActionsTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <ActionsTableSkeleton />
