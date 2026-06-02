export default {
  title: 'Runbooks/RunbooksTable',
}

import { RunbooksTable, RunbooksTableSkeleton, type TRunbookRow } from './RunbooksTable'

const mockRows: TRunbookRow[] = Array.from({ length: 3 }, (_, i) => ({
  runbookId: `runbook-${i + 1}`,
  runbookName: `rotate-secrets-${i + 1}`,
  description: <span className="text-sm text-gray-500">Rotates API keys and secrets for the install.</span>,
  labels: <span className="text-sm text-gray-500">production</span>,
  lastUpdated: <span className="text-sm text-gray-500">3 days ago</span>,
  href: `/org-1/apps/app-1/runbooks/runbook-${i + 1}`,
}))

export const Default = () => (
  <RunbooksTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <RunbooksTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <RunbooksTableSkeleton />
