export default {
  title: 'Playground / Table v2 / Branch installs',
}

import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { CloudRegion } from '@/components/common/CloudRegion'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import {
  INSTALL_STATUS_AXES,
  InstallStatusAxis,
} from '@/components/installs/InstallStatusSummary/InstallStatusSummary'
import type { TCloudPlatform, TInstallStatus, TInstallStatusAxis } from '@/types'
import { StoryFrame, hoursAgo, useMockPage } from './story-utils'
import { TableV2 } from './TableV2'

type TMockBranchInstall = {
  createdAt: string
  createdBy: string
  group?: string
  id: string
  labels: Record<string, string>
  name: string
  platform: TCloudPlatform
  region: string
  status: TInstallStatus
  updatedAt: string
}

const axis = (
  status: string,
  description: string
): TInstallStatusAxis => ({ status, status_human_description: description })

const STATUS_SCENARIOS: TInstallStatus[] = [
  {
    deployments: axis('active', 'All deployed'),
    resources: axis('healthy', 'All healthy'),
    health_checks: axis('healthy', 'All healthy'),
  },
  {
    deployments: axis('pending', 'In progress'),
    resources: axis('healthy', 'All healthy'),
    health_checks: axis('healthy', 'All healthy'),
  },
  {
    deployments: axis('error', '1 deployment failed'),
    resources: axis('degraded', 'Degraded'),
    health_checks: axis('healthy', 'All healthy'),
  },
  {
    deployments: axis('active', 'All deployed'),
    resources: axis('unhealthy', 'Unhealthy'),
    health_checks: axis('unhealthy', 'Unhealthy'),
  },
  {
    deployments: axis('pending', 'Not deployed'),
    resources: axis('unknown', 'No resources'),
    health_checks: axis('unknown', 'No health checks'),
  },
]

const CUSTOMERS = [
  'northwind',
  'globex',
  'initech',
  'umbrella',
  'hooli',
  'stark',
  'wayne',
  'wonka',
]
const GROUPS = ['production', 'canary', 'staging', undefined]
const PLATFORMS: { platform: TCloudPlatform; region: string }[] = [
  { platform: 'aws', region: 'us-east-1' },
  { platform: 'aws', region: 'eu-west-1' },
  { platform: 'azure', region: 'eastus' },
  { platform: 'gcp', region: 'us-central1' },
  { platform: 'azure', region: 'westeurope' },
]

const mockInstalls: TMockBranchInstall[] = Array.from(
  { length: 34 },
  (_, i) => {
    const customer = CUSTOMERS[i % CUSTOMERS.length]
    const target = PLATFORMS[(i * 3) % PLATFORMS.length]
    return {
      createdAt: hoursAgo(24 * (40 + i)),
      createdBy:
        i % 3 === 0 ? 'ops-bot@example.com' : `admin${(i % 4) + 1}@example.com`,
      group: GROUPS[i % GROUPS.length],
      id: `inst${String(i + 1).padStart(3, '0')}example`,
      labels:
        i % 4 === 0
          ? {}
          : i % 4 === 1
            ? { tier: 'enterprise' }
            : { tier: 'standard', env: i % 2 ? 'prod' : 'staging' },
      name: `${customer}-prod-${i + 1}`,
      platform: target.platform,
      region: target.region,
      status: STATUS_SCENARIOS[(i * 2) % STATUS_SCENARIOS.length],
      updatedAt: hoursAgo(2 + i * 7),
    }
  }
)

const ActivityCell = ({ install }: { install: TMockBranchInstall }) => (
  <ContextTooltip
    position="top"
    width="w-64"
    items={[
      {
        id: 'created-by',
        title: install.createdBy,
        subtitle: install.createdBy.startsWith('ops-bot')
          ? 'via API / CLI'
          : 'via Dashboard',
        leftContent: <Icon variant="UserIcon" size={16} />,
      },
      {
        id: 'created',
        title: 'Created',
        subtitle: (
          <Time
            variant="label"
            time={install.createdAt}
            format="long-datetime"
          />
        ),
        leftContent: <Icon variant="PlusCircleIcon" size={16} />,
      },
      {
        id: 'updated',
        title: 'Updated',
        subtitle: (
          <Time
            variant="label"
            time={install.updatedAt}
            format="long-datetime"
          />
        ),
        leftContent: <Icon variant="ClockCounterClockwiseIcon" size={16} />,
      },
    ]}
  >
    <span className="inline-flex items-center gap-1.5 cursor-default">
      <Time time={install.updatedAt} variant="subtext" format="relative" />
      <Icon variant="InfoIcon" size={12} theme="neutral" />
    </span>
  </ContextTooltip>
)

const columns: ColumnDef<TMockBranchInstall>[] = [
  {
    accessorKey: 'name',
    header: 'Install name',
    enableHiding: false,
    size: 220,
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link
            href={`/org-1/installs/${info.row.original.id}`}
            variant="inline"
          >
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.id}</ID>
      </span>
    ),
  },
  {
    id: 'statusAxes',
    accessorFn: (row) => row.status.deployments?.status,
    header: 'Statuses',
    enableSorting: false,
    size: 190,
    cell: (info) => (
      <div className="flex w-max flex-col items-start gap-1.5">
        {INSTALL_STATUS_AXES.map(({ key, label }) => (
          <span key={key} title={label}>
            <InstallStatusAxis axis={info.row.original.status[key]} />
          </span>
        ))}
      </div>
    ),
  },
  {
    accessorKey: 'region',
    header: 'Region',
    size: 190,
    meta: { defaultHidden: true },
    cell: (info) => {
      const { platform, region } = info.row.original
      return (
        <CloudRegion
          variant="subtext"
          platform={platform}
          region={region}
          location={region}
        />
      )
    },
  },
  {
    accessorKey: 'labels',
    header: 'Labels',
    enableSorting: false,
    size: 200,
    cell: (info) => {
      const labels = info.getValue() as TMockBranchInstall['labels']
      const keys = Object.keys(labels).sort()
      if (!keys.length) return <Icon variant="MinusIcon" />
      return (
        <span className="flex flex-wrap gap-1">
          {keys.map((key) => (
            <LabelBadge
              key={key}
              size="sm"
              labelKey={key}
              labelValue={labels[key]}
            />
          ))}
        </span>
      )
    },
  },
  {
    accessorKey: 'group',
    header: 'Group',
    enableSorting: false,
    size: 130,
    cell: (info) => (
      <Text variant="subtext" theme="neutral" family="mono">
        {(info.getValue() as string | undefined) || '—'}
      </Text>
    ),
  },
  {
    id: 'activity',
    accessorKey: 'updatedAt',
    header: 'Activity',
    size: 140,
    cell: (info) => <ActivityCell install={info.row.original} />,
  },
  {
    id: 'action',
    header: '',
    size: 56,
    enableSorting: false,
    enableHiding: false,
    cell: () => (
      <Button aria-label="Manage install" size="sm" variant="ghost">
        <Icon variant="DotsThreeVerticalIcon" />
      </Button>
    ),
  },
]

const PAGE_SIZE = 20

const matchesInstall = (install: TMockBranchInstall, query: string) =>
  install.name.includes(query) ||
  install.id.includes(query) ||
  (install.group?.includes(query) ?? false)

const emptyStateProps = {
  emptyTitle: 'No installs managed by this branch',
  emptyMessage:
    "Create an install and assign it to one of this branch's deployment groups.",
}

export const Default = () => {
  const { data, pagination } = useMockPage(
    mockInstalls,
    PAGE_SIZE,
    matchesInstall
  )
  return (
    <StoryFrame note="Mirrors the installs table on a branch page (scope branch with status axes). The app and branch columns are left out because the page already implies them. Region starts hidden because the current table only shows it on very wide screens.">
      <TableV2
        tableId="playground-branch-installs"
        columns={columns}
        data={data}
        pagination={pagination}
        searchPlaceholder="Search by name or ID..."
        emptyStateProps={emptyStateProps}
      />
    </StoryFrame>
  )
}

export const Loading = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-branch-installs"
      columns={columns}
      data={[]}
      isLoading
      pagination={{ limit: PAGE_SIZE, offset: 0 }}
      searchPlaceholder="Search by name or ID..."
    />
  </StoryFrame>
)

export const Empty = () => (
  <StoryFrame>
    <TableV2
      tableId="playground-branch-installs"
      columns={columns}
      data={[]}
      searchPlaceholder="Search by name or ID..."
      emptyStateProps={emptyStateProps}
    />
  </StoryFrame>
)
