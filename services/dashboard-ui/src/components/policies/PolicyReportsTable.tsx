'use client'

import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportPanelButton } from '@/components/policies/PolicyReportPanel'
import type { TPolicyReport } from '@/types'

type TPolicyReportRow = {
  id: string
  componentName: string | null
  ownerType: string
  denyCount: number
  warnCount: number
  passCount: number
  evaluatedAt: string
  status: string
  report: TPolicyReport
}

function formatOwnerType(ownerType: string): {
  label: string
  theme: 'info' | 'brand' | 'neutral'
} {
  switch (ownerType) {
    case 'install_deploys':
      return { label: 'Deploy', theme: 'info' }
    case 'install_sandbox_runs':
      return { label: 'Sandbox', theme: 'brand' }
    case 'component_builds':
      return { label: 'Build', theme: 'neutral' }
    default:
      return { label: ownerType, theme: 'neutral' }
  }
}

function getStatusBadge(row: TPolicyReportRow) {
  if (row.denyCount > 0) {
    return (
      <Badge theme="error" size="sm">
        <Icon variant="XCircle" size={10} />
        Denied
      </Badge>
    )
  }
  if (row.warnCount > 0) {
    return (
      <Badge theme="warn" size="sm">
        <Icon variant="Warning" size={10} />
        Warning
      </Badge>
    )
  }
  return (
    <Badge theme="success" size="sm">
      <Icon variant="CheckCircle" size={10} />
      Passed
    </Badge>
  )
}

function parsePolicyReportsToTableData(
  reports: TPolicyReport[]
): TPolicyReportRow[] {
  return reports.map((report) => ({
    id: report.id || '',
    componentName: report.component_name || null,
    ownerType: report.owner_type || '',
    denyCount: report.deny_count || 0,
    warnCount: report.warn_count || 0,
    passCount: report.pass_count || 0,
    evaluatedAt: report.evaluated_at || '',
    status: report.status?.status || '',
    report,
  }))
}

export const policyReportsTableColumns: ColumnDef<TPolicyReportRow>[] = [
  {
    accessorKey: 'componentName',
    header: 'Component',
    cell: (info) => {
      const name = info.getValue() as string | null
      return (
        <span>
          <Text variant="body">{name || 'Sandbox'}</Text>
          <ID>{info.row.original.id}</ID>
        </span>
      )
    },
    enableSorting: true,
  },
  {
    accessorKey: 'ownerType',
    header: 'Type',
    cell: (info) => {
      const { label, theme } = formatOwnerType(info.getValue() as string)
      return (
        <Badge theme={theme} size="sm">
          {label}
        </Badge>
      )
    },
  },
  {
    accessorKey: 'status',
    header: 'Status',
    cell: (info) => getStatusBadge(info.row.original),
  },
  {
    accessorKey: 'denyCount',
    header: 'Denies',
    cell: (info) => {
      const count = info.getValue() as number
      return count > 0 ? (
        <Text variant="body" className="text-red-600 dark:text-red-400">
          {count}
        </Text>
      ) : (
        <Text variant="body" theme="neutral">
          0
        </Text>
      )
    },
  },
  {
    accessorKey: 'warnCount',
    header: 'Warnings',
    cell: (info) => {
      const count = info.getValue() as number
      return count > 0 ? (
        <Text variant="body" className="text-orange-600 dark:text-orange-400">
          {count}
        </Text>
      ) : (
        <Text variant="body" theme="neutral">
          0
        </Text>
      )
    },
  },
  {
    accessorKey: 'evaluatedAt',
    header: 'Evaluated',
    cell: (info) => <Time time={info.getValue() as string} format="relative" />,
    enableSorting: true,
  },
]

const createActionsColumn = (orgId: string): ColumnDef<TPolicyReportRow> => ({
  id: 'actions',
  header: '',
  cell: (info) => (
    <PolicyReportPanelButton report={info.row.original.report} orgId={orgId} />
  ),
})

export const PolicyReportsTable = ({
  reports,
  orgId,
  installId,
}: {
  reports: TPolicyReport[]
  orgId: string
  installId: string
}) => {
  const data = parsePolicyReportsToTableData(reports)
  const columns = [...policyReportsTableColumns, createActionsColumn(orgId)]

  return (
    <Table<TPolicyReportRow>
      columns={columns}
      data={data}
      emptyStateProps={{
        emptyMessage:
          'Policy evaluations will appear here after deploys or sandbox runs.',
        emptyTitle: 'No policy evaluations',
      }}
      pagination={{
        limit: data.length || 10,
        offset: 0,
        hasNext: false,
      }}
    />
  )
}
