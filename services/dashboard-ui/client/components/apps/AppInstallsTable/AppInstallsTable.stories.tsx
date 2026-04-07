export default {
  title: 'Apps/AppInstallsTable',
}

import { AppInstallsTable, AppInstallsTableSkeleton, type InstallRow } from './AppInstallsTable'
import { Button } from '@/components/common/Button'

const mockRows: InstallRow[] = Array.from({ length: 3 }, (_, i) => ({
  actionHref: `/org-1/installs/install-${i + 1}`,
  installId: `install-${i + 1}`,
  name: `Install ${i + 1}`,
  nameHref: `/org-1/installs/install-${i + 1}`,
  region: <span>us-east-1</span>,
  statuses: <span>active</span>,
  platform: <span>AWS</span>,
}))

export const Default = () => (
  <AppInstallsTable
    data={mockRows}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <AppInstallsTable
    data={[]}
    isLoading={false}
    emptyAction={<Button>Create install</Button>}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <AppInstallsTableSkeleton />
