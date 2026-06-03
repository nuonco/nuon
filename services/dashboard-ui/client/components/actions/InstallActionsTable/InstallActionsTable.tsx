import type { ReactNode } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Duration } from '@/components/common/Duration'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { StatusWithDescription } from '@/components/common/StatusWithDescription'
import { Table } from '@/components/common/Table'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TActionConfigTriggerType, TInstallAction } from '@/types'
import { toSentenceCase } from '@/utils/string-utils'
import { ActionTriggerType } from '../ActionTriggerType'

export type InstallActionRow = {
  actionId: string
  actionName: string
  actionStatus: ReactNode
  actionTrigger: ReactNode
  actionRunDatetime: ReactNode
  actionRunDuration: ReactNode
  labels: ReactNode
  href: string
}

export function parseInstallActionsLatestRunsToTableData(
  actionsWithRuns: TInstallAction[],
  orgId: string,
  installId: string
): InstallActionRow[] {
  return actionsWithRuns.map((actionWithRuns) => {
    const basePath = `/${orgId}/installs/${installId}`
    const recentRun = actionWithRuns?.runs?.at(0)

    return {
      actionId: actionWithRuns.action_workflow_id,
      actionName: actionWithRuns.action_workflow?.name,
      actionRunDatetime: recentRun ? (
        <Text flex className="gap-2">
          <Icon variant="CalendarBlankIcon" />
          <Time
            time={recentRun?.created_at}
            format="relative"
            variant="subtext"
          />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionRunDuration: recentRun ? (
        <Text flex className="gap-2">
          <Icon variant="TimerIcon" />
          <Duration nanoseconds={recentRun?.execution_time} variant="subtext" />
        </Text>
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionStatus: recentRun ? (
        <StatusWithDescription
          statusProps={{ status: recentRun?.status_v2?.status }}
          tooltipProps={{
            position: 'top',
            tipContent: toSentenceCase(
              recentRun?.status_v2?.status_human_description
            ),
          }}
        />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      actionTrigger: recentRun ? (
        <ActionTriggerType
          componentName={recentRun?.run_env_vars?.COMPONENT_NAME}
          componentPath={`${basePath}/components/${recentRun?.run_env_vars?.COMPONENT_ID}`}
          triggerType={recentRun?.triggered_by_type as TActionConfigTriggerType}
        />
      ) : (
        <Icon variant="MinusIcon" />
      ),
      labels: (() => {
        const lbls = actionWithRuns.action_workflow?.labels
        if (!lbls || Object.keys(lbls).length === 0) return null
        return (
          <span className="flex flex-wrap gap-1">
            {Object.keys(lbls)
              .sort()
              .map((k) => (
                <LabelBadge key={k} labelKey={k} labelValue={lbls[k]} size="sm" />
              ))}
          </span>
        )
      })(),
      href: `${basePath}/actions/${actionWithRuns.action_workflow_id}`,
    }
  })
}

const columns: ColumnDef<InstallActionRow>[] = [
  {
    accessorKey: 'actionName',
    header: 'Action',
    cell: (info) => (
      <span>
        <Text variant="body">
          <Link href={info.row.original.href}>{info.getValue() as string}</Link>
        </Text>
        <ID>{info.row.original.actionId as string}</ID>
      </span>
    ),
    enableSorting: true,
  },
  {
    accessorKey: 'actionRunDatetime',
    header: 'Last run',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'actionRunDuration',
    header: 'Duration',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'actionTrigger',
    header: 'Recent trigger',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'labels',
    header: 'Labels',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    accessorKey: 'actionStatus',
    header: 'Status',
    cell: (info) => info.getValue() as ReactNode,
  },
  {
    enableSorting: false,
    accessorKey: 'href',
    id: 'action',
    header: '',
    cell: (info) => (
      <Text>
        <Link className="text-left" href={info.getValue() as string}>
          View <Icon variant="CaretRightIcon" />
        </Link>
      </Text>
    ),
  },
]

interface IInstallActionsTable {
  data: InstallActionRow[]
  isLoading?: boolean
  filterActions?: ReactNode
  pagination: { hasNext?: boolean; offset: number; limit: number }
}

export const InstallActionsTable = ({
  data,
  isLoading,
  filterActions,
  pagination,
}: IInstallActionsTable) => {
  return (
    <Table<InstallActionRow>
      columns={columns}
      data={data}
      isLoading={isLoading}
      filterActions={filterActions}
      emptyStateProps={{
        variant: 'actions',
        emptyMessage:
          'Save time by configuring your actions. Check out our resources.',
        emptyTitle: 'No actions yet',
        action: (
          <Link href="https://docs.nuon.co/concepts/actions" isExternal>
            Learn more <Icon size="14" variant="ArrowSquareOutIcon" />
          </Link>
        ),
      }}
      pagination={pagination}
      searchPlaceholder="Search by name or ID..."
    />
  )
}

export const InstallActionsTableSkeleton = () => {
  return <TableSkeleton columns={columns} skeletonRows={5} />
}
