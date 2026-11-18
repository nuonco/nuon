export default {
  title: 'Playground / Table v2 / Apps',
}

import type { ColumnDef } from '@tanstack/react-table'
import { Button } from '@/components/common/Button'
import { CloudPlatform } from '@/components/common/CloudPlatform'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import type { TAppRow } from '@/components/apps/AppsTable/AppsTable'
import type { TCloudPlatform } from '@/types'
import { StoryFrame, useMockPage } from './story-utils'
import { TableV2 } from '@/components/common/TableV2'

const APP_NAMES = [
  'payments',
  'analytics',
  'search',
  'billing',
  'notifications',
  'inventory',
  'checkout',
  'reporting',
  'identity',
  'catalog',
  'messaging',
  'scheduler',
  'warehouse',
]
const PLATFORMS: TCloudPlatform[] = ['aws', 'azure', 'gcp', 'aws', 'unknown']

const mockApps: TAppRow[] = Array.from({ length: 26 }, (_, i) => {
  const base = APP_NAMES[i % APP_NAMES.length]
  const name = i < APP_NAMES.length ? `acme-${base}` : `acme-${base}-eu`
  const id = `app${String(i + 1).padStart(3, '0')}example`
  const platform = PLATFORMS[i % PLATFORMS.length]
  const hasSandbox = i % 5 !== 3
  const sandboxName = `acme-example/sandbox-${platform === 'unknown' ? 'aws' : platform}`
  return {
    actionHref: `/org-1/apps/${id}`,
    appId: id,
    configVersion: i % 7 === 6 ? 0 : (i * 3) % 14 + 1,
    defaultBranch: 'main',
    name,
    nameHref: `/org-1/apps/${id}`,
    platform,
    sandboxHref: hasSandbox ? `https://example.com/${sandboxName}` : undefined,
    sandboxName: hasSandbox ? sandboxName : undefined,
  } as TAppRow
})

const columns: ColumnDef<TAppRow>[] = [
  {
    accessorKey: 'name',
    header: 'App name',
    enableHiding: false,
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.nameHref} variant="inline">
            {info.getValue() as string}
          </Link>
        </Text>
        <ID>{info.row.original.appId}</ID>
      </span>
    ),
  },
  {
    accessorKey: 'configVersion',
    header: 'Config version',
    cell: (info) => (
      <Text family="mono" theme="neutral">
        {info.getValue<number>() === 0
          ? 'No config'
          : info.getValue<number>()}
      </Text>
    ),
  },
  {
    accessorKey: 'sandboxName',
    header: 'Sandbox',
    cell: (info) =>
      info.row.original.sandboxHref ? (
        <Text>
          <Link
            href={info.row.original.sandboxHref}
            isExternal
            variant="inline"
          >
            {info.getValue() as string}
          </Link>
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
  },
  {
    accessorKey: 'platform',
    header: 'Platform',
    cell: (info) => (
      <CloudPlatform
        displayVariant="icon-only"
        iconSize="24"
        platform={info.getValue() as TCloudPlatform}
        colorVariant="color"
      />
    ),
  },
  {
    id: 'action',
    accessorKey: 'actionHref',
    header: '',
    enableSorting: false,
    enableHiding: false,
    cell: (info) => (
      <Text>
        <Link
          className="text-left"
          href={info.getValue() as string}
          variant="inline"
        >
          View app
        </Link>
      </Text>
    ),
  },
]

const PAGE_SIZE = 20

const matchesApp = (app: TAppRow, query: string) =>
  app.name.includes(query) || app.appId.includes(query)

const emptyStateProps = {
  variant: 'app' as const,
  emptyTitle: 'No apps yet',
  emptyMessage:
    'An app is the configuration that gets deployed into your customers cloud accounts.',
  action: (
    <Button variant="primary">
      <Icon variant="PlusIcon" />
      Create app
    </Button>
  ),
}

export const Default = () => {
  const { data, pagination } = useMockPage(mockApps, PAGE_SIZE, matchesApp)
  return (
    <StoryFrame note="Mirrors the org apps table. Platform is an icon-only column, which the grid shows as a labelled value. Apps without a sandbox repo or config show the same placeholders as today.">
      <TableV2
        tableId="playground-apps"
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
      tableId="playground-apps"
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
      tableId="playground-apps"
      columns={columns}
      data={[]}
      searchPlaceholder="Search by name or ID..."
      emptyStateProps={emptyStateProps}
    />
  </StoryFrame>
)
