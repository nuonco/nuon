export default {
  title: 'Runbooks/InstallRunbooksTable',
}

import { InstallRunbooksTable, InstallRunbooksTableSkeleton, type TInstallRunbookRow } from './InstallRunbooksTable'

const mockRows: TInstallRunbookRow[] = Array.from({ length: 3 }, (_, i) => ({
  runbookId: `runbook-${i + 1}`,
  runbookName: `rotate-secrets-${i + 1}`,
  description: <span className="text-sm text-gray-500">Rotates API keys and credentials.</span>,
  labels: <span className="text-sm text-gray-500">production</span>,
  href: `/org-1/installs/install-1/runbooks/runbook-${i + 1}`,
}))

export const Default = () => (
  <InstallRunbooksTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <InstallRunbooksTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <InstallRunbooksTableSkeleton />
