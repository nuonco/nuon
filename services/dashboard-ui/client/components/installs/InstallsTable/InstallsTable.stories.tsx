export default {
  title: 'Installs/InstallsTable',
}

import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import { AirgapBadge } from '@/components/installs/AirgapBadge'
import {
  InstallsTable,
  InstallsTableSkeleton,
  type InstallRow,
} from './InstallsTable'

const mockRows: InstallRow[] = Array.from({ length: 5 }, (_, i) => ({
  name: `prod-acme-${i + 1}`,
  nameHref: `/org-1/installs/install-${i + 1}`,
  installId: `install-${i + 1}`,
  appName: 'My BYOC App',
  appHref: `/org-1/apps/app-1`,
  statuses: <span className="text-sm text-foreground-muted">active</span>,
  region: <span className="text-sm text-foreground-muted">us-west-2</span>,
  platform: <span className="text-sm text-foreground-muted">AWS</span>,
  labels:
    i % 2 === 0 ? <span className="text-xs font-mono">env: prod</span> : null,
  branch: null,
  activity: <span className="text-sm text-foreground-muted">{i + 1}h ago</span>,
  updatedAt: new Date(Date.now() - (i + 1) * 60 * 60 * 1000).toISOString(),
  action: (
    <Button size="sm" variant="ghost">
      Manage
    </Button>
  ),
}))

const airgapRow: InstallRow = {
  name: 'acme-airgapped',
  nameHref: '/org-1/installs/install-airgap',
  installId: 'install-airgap',
  appName: 'My BYOC App',
  appHref: '/org-1/apps/app-1',
  statuses: <AirgapBadge />,
  region: <Icon variant="MinusIcon" />,
  platform: <Icon variant="MinusIcon" />,
  labels: null,
  branch: null,
  activity: <span className="text-sm text-foreground-muted">2h ago</span>,
  updatedAt: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
  action: (
    <Button size="sm" variant="ghost">
      Manage
    </Button>
  ),
}

export const Default = () => (
  <InstallsTable
    data={mockRows}
    isLoading={false}
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<Button variant="primary">Create install</Button>}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const WithAirgapInstall = () => (
  <InstallsTable
    data={[airgapRow, ...mockRows]}
    isLoading={false}
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<Button variant="primary">Create install</Button>}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <InstallsTable
    data={[]}
    isLoading={false}
    emptyStateAction={<Button>Create install</Button>}
    filterActions={<Button variant="primary">Create install</Button>}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <InstallsTableSkeleton />
