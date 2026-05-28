import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Dropdown } from '@/components/common/Dropdown'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Menu } from '@/components/common/Menu'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { RunRunbookButton } from '@/components/runbooks/RunRunbook/RunRunbook'
import type { TInstallRunbook } from '@/lib/ctl-api/installs/runbooks'

export type TInstallRunbookRow = {
  runbookId: string
  runbookName: string
  description: ReactNode
  labels: ReactNode
  lastRun: ReactNode
  href: string
  latestRunHref: string | null
  installRunbook: TInstallRunbook
}

export function parseInstallRunbooksToTableData(
  runbooks: TInstallRunbook[],
  orgId: string,
  installId: string
): TInstallRunbookRow[] {
  return runbooks.map((ir) => {
    const basePath = `/${orgId}/installs/${installId}`
    const runbook = ir.runbook
    return {
      runbookId: ir.runbook_id ?? ir.id,
      runbookName: runbook?.name ?? '',
      description: runbook?.description ? (
        <Text variant="subtext" theme="neutral">
          {runbook.description}
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels:
        runbook?.labels && Object.keys(runbook.labels).length > 0 ? (
          <span className="flex flex-wrap gap-1">
            {Object.keys(runbook.labels)
              .sort()
              .map((k) => (
                <Badge key={k} variant="code" size="sm" theme="neutral">
                  {k}: {runbook.labels[k]}
                </Badge>
              ))}
          </span>
        ) : (
          <Icon variant="MinusIcon" />
        ),
      lastRun: ir.runs?.[0] ? (
        <Text flex className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={ir.runs[0].created_at}
            format="relative"
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      href: `${basePath}/runbooks/${ir.runbook_id ?? ir.id}`,
      latestRunHref: (() => {
        const latestRun = ir.runs?.[0]
        const workflowId = latestRun?.install_workflow_id ?? latestRun?.install_workflow?.id
        return workflowId ? `${basePath}/workflows/${workflowId}` : null
      })(),
      installRunbook: ir,
    }
  })
}

const columns: ColumnDef<TInstallRunbookRow>[] = [
  {
    accessorKey: 'runbookName',
    header: 'Runbook',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
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
    accessorKey: 'lastRun',
    header: 'Last run',
    cell: (info) => info.getValue() as ReactNode,
    enableSorting: false,
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'action',
    header: '',
    cell: (info) => (
      <Dropdown
        alignment="right"
        buttonText=""
        buttonClassName="!p-1"
        icon={<Icon variant="DotsThreeVerticalIcon" />}
        id={info.row.original.runbookId}
        variant="ghost"
      >
        <Menu>
          <RunRunbookButton
            installRunbook={info.row.original.installRunbook}
            isMenuButton
          >
            Run runbook
          </RunRunbookButton>
          {info.row.original.latestRunHref ? (
            <Button href={info.row.original.latestRunHref} isMenuButton>
              Latest run
              <Icon variant="CaretRightIcon" />
            </Button>
          ) : null}
          <Button href={info.row.original.href} isMenuButton>
            View details
            <Icon variant="CaretRightIcon" />
          </Button>
        </Menu>
      </Dropdown>
    ),
  },
]

interface IInstallRunbooksTable {
  data: TInstallRunbookRow[]
  isLoading?: boolean
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const InstallRunbooksTable = ({ data, isLoading, pagination }: IInstallRunbooksTable) => {
  return (
    <Table<TInstallRunbookRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      emptyStateProps={{
        variant: 'actions',
        emptyTitle: 'No runbooks yet',
        emptyMessage: 'Runbooks let you run operational procedures on this install.',
      }}
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const InstallRunbooksTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
