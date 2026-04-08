export default {
  title: 'Stacks/InstallStacksTable',
}

import { InstallStacksTable, InstallStacksTableSkeleton, type TInstallStackRow } from './InstallStacksTable'
import { Status } from '@/components/common/Status'

const mockData: TInstallStackRow[] = Array.from({ length: 3 }, (_, i) => ({
  versionId: `ver-${i + 1}`,
  appConfigId: `cfg-${i + 1}`,
  appStackConfigHref: `/org-1/apps/app-1`,
  status: <Status variant="badge" status="active" />,
  runs: `${i + 1}`,
  createdAt: new Date(Date.now() - i * 86400000).toISOString(),
  more: <span>details</span>,
}))

export const Default = () => (
  <InstallStacksTable
    data={mockData}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
  />
)

export const Empty = () => (
  <InstallStacksTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 10 }}
  />
)

export const Skeleton = () => <InstallStacksTableSkeleton />
