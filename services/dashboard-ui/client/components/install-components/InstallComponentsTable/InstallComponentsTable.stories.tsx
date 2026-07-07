export default {
  title: 'Install Components/InstallComponentsTable',
}

import { InstallComponentsTable, type InstallComponentRow } from './InstallComponentsTable'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { Text } from '@/components/common/Text'

const mockRows: InstallComponentRow[] = [
  {
    componentId: 'comp-abc123',
    componentName: 'web-server',
    componentType: <Text variant="subtext">helm_chart</Text>,
    toggleStatus: <Badge size="sm" theme="success">Enabled</Badge>,
    deployStatus: <Text variant="subtext">active</Text>,
    driftStatus: <Text variant="subtext">-</Text>,
    href: '/org1/installs/inst1/components/comp-abc123',
    action: <div />,
    dependencies: <Text variant="subtext">-</Text>,
    labels: <Text variant="subtext">tier: application</Text>,
  },
  {
    componentId: 'comp-def456',
    componentName: 'database',
    componentType: <Text variant="subtext">terraform_module</Text>,
    toggleStatus: <Icon variant="MinusIcon" />,
    deployStatus: <Text variant="subtext">active</Text>,
    driftStatus: <Text variant="subtext">drifted</Text>,
    href: '/org1/installs/inst1/components/comp-def456',
    action: <div />,
    dependencies: <Text variant="subtext">-</Text>,
    labels: null,
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
