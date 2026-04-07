export default {
  title: 'Install Components/InstallComponentsTable',
}

import { InstallComponentsTable, type InstallComponentRow } from './InstallComponentsTable'
import { Text } from '@/components/common/Text'

const mockRows: InstallComponentRow[] = [
  {
    componentId: 'comp-abc123',
    componentName: 'web-server',
    componentType: <Text variant="subtext">helm_chart</Text>,
    deployStatus: <Text variant="subtext">active</Text>,
    driftStatus: <Text variant="subtext">-</Text>,
    href: '/org1/installs/inst1/components/comp-abc123',
    action: <div />,
    dependencies: <Text variant="subtext">-</Text>,
  },
  {
    componentId: 'comp-def456',
    componentName: 'database',
    componentType: <Text variant="subtext">terraform_module</Text>,
    deployStatus: <Text variant="subtext">active</Text>,
    driftStatus: <Text variant="subtext">drifted</Text>,
    href: '/org1/installs/inst1/components/comp-def456',
    action: <div />,
    dependencies: <Text variant="subtext">-</Text>,
  },
]

const mockPagination = { hasNext: false, offset: 0, limit: 10 }

export const Default = () => (
  <InstallComponentsTable
    data={mockRows}
    filterActions={<div />}
    pagination={mockPagination}
    isLoading={false}
  />
)

export const Loading = () => (
  <InstallComponentsTable
    data={[]}
    filterActions={<div />}
    pagination={mockPagination}
    isLoading={true}
  />
)

export const Empty = () => (
  <InstallComponentsTable
    data={[]}
    filterActions={<div />}
    pagination={mockPagination}
    isLoading={false}
  />
)
