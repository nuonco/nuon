import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RemovedFromAppConfigBadge } from '@/components/installs/RemovedFromAppConfig'
import type { TRunbook } from '@/lib/ctl-api/apps/runbooks'

export type TRunbooksTableScope = 'app' | 'install'

export type TRunbookRow = {
  runbookId: string
  runbookName: string
  description: ReactNode
  labels: ReactNode
  lastUpdated: ReactNode
  href: string
  lastRun?: ReactNode
  actions?: ReactNode
  removed?: boolean
}

export function parseRunbooksToTableData(
  runbooks: TRunbook[],
  orgId: string,
  appId: string,
  labelColors?: Record<string, string>,
  branchId?: string
): TRunbookRow[] {
  return runbooks.map((runbook) => {
    const basePath = branchId
      ? `/${orgId}/apps/${appId}/branches/${branchId}`
      : `/${orgId}/apps/${appId}`
    return {
      runbookId: runbook.id,
      runbookName: runbook.name,
      description: runbook.description ? (
        <Text variant="subtext" theme="neutral">
          {runbook.description}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels:
        runbook.labels && Object.keys(runbook.labels).length > 0 ? (
          <span className="flex flex-wrap gap-1">
            {Object.keys(runbook.labels)
              .sort()
              .map((k) => (
                <LabelBadge
                  key={k}
                  labelKey={k}
                  labelValue={runbook.labels[k]}
                  size="sm"
                  customColor={labelColors?.[k]}
                />
              ))}
          </span>
        ) : (
          <Icon variant="MinusIcon" />
        ),
      lastUpdated: runbook.updated_at ? (
        <Text flex className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time time={runbook.updated_at} format="relative" variant="subtext" />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      href: `${basePath}/runbooks/${runbook.id}`,
    }
  })
}

const columns: ColumnDef<TRunbookRow>[] = [
  {
    accessorKey: 'runbookName',
    header: 'Runbook',
    cell: (info) => (
      <span>
        <Text variant="body" flex className="items-center gap-2">
          <Link href={info.row.original.href} variant="inline">
            {info.getValue() as string}
          </Link>
          {info.row.original.removed ? (
            <RemovedFromAppConfigBadge kind="runbook" />
          ) : null}
        </Text>
        <ID>{info.row.original.runbookId}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'description',
    header: 'Description',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'labels',
    header: 'Labels',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    accessorKey: 'lastUpdated',
    header: 'Last updated',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
]

const lastRunColumn: ColumnDef<TRunbookRow> = {
  accessorKey: 'lastRun',
  header: 'Last run',
  cell: (info) => info.getValue() as ReactNode,
  enableSorting: false,
}

const actionsColumn: ColumnDef<TRunbookRow> = {
  enableSorting: false,
  accessorKey: 'actions',
  id: 'action',
  header: '',
  cell: (info) => info.getValue() as ReactNode,
}

interface IRunbooksTable {
  data: TRunbookRow[]
  filterActions?: ReactNode
  isLoading: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
  scope?: TRunbooksTableScope
}

export const RunbooksTable = ({
  data,
  filterActions,
  isLoading,
  pagination,
  scope = 'app',
}: IRunbooksTable) => {
  const scopedColumns =
    scope === 'install' ? [...columns, lastRunColumn, actionsColumn] : columns

  return (
    <Table<TRunbookRow>
      columns={scopedColumns}
      data={data}
      filterActions={filterActions}
      isLoading={isLoading}
      emptyStateProps={{
        variant: 'actions',
        emptyTitle: 'No runbooks yet',
        emptyMessage:
          'Runbooks let you define operational procedures for your installs. Add runbooks to your app config and sync to see them here.',
      }}
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}
