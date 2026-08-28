import { useMemo } from 'react'
import type { ColumnDef } from '@tanstack/react-table'
import { Badge } from '@/components/common/Badge'
import { Status } from '@/components/common/Status'
import { Table } from '@/components/common/Table'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PolicyReportPanel } from '@/components/policies/PolicyReportPanel'
import { panelTriggerClass } from '@/components/surfaces/panel-trigger'
import type { TPolicyReport } from '@/types'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

export interface IPolicyReportRow {
  key: string
  report: TPolicyReport
  history: TPolicyReport[]
}

const OWNER_TYPE_LABELS: Record<string, { label: string; theme: 'info' | 'brand' | 'neutral' }> = {
  install_deploys: { label: 'Deploy', theme: 'info' },
  install_sandbox_runs: { label: 'Sandbox', theme: 'brand' },
  component_builds: { label: 'Build', theme: 'neutral' },
}

function getGroupKey(report: TPolicyReport): string {
  const ownerType = report?.owner_type ?? 'unknown'
  return report?.component_id ? `${ownerType}:${report.component_id}` : ownerType
}

function getReportTime(report: TPolicyReport): number {
  return report?.evaluated_at ? new Date(report.evaluated_at).getTime() : 0
}

export function groupPolicyReports(
  reports: TPolicyReport[]
): IPolicyReportRow[] {
  const groups = new Map<string, TPolicyReport[]>()
  for (const report of reports ?? []) {
    const key = getGroupKey(report)
    groups.set(key, [...(groups.get(key) ?? []), report])
  }

  return Array.from(groups.entries())
    .map(([key, items]) => {
      const sorted = [...items].sort((a, b) => getReportTime(b) - getReportTime(a))
      const [latest, ...history] = sorted
      return { key, report: latest, history }
    })
    .sort((a, b) => getReportTime(b.report) - getReportTime(a.report))
}

export const reportSubject = (report: TPolicyReport): string =>
  report?.component_name ? `Component - ${report.component_name}` : 'Sandbox'

export const PolicyReportsTable = ({
  reports,
  orgId,
  policyNameMap,
  isLoading,
  currentStatus,
  currentOwnerType,
}: {
  reports: TPolicyReport[]
  orgId: string
  policyNameMap: Map<string, string>
  isLoading?: boolean
  currentStatus?: TPolicyReportStatus
  currentOwnerType?: TPolicyReportOwnerType
}) => {
  const hasActiveFilters = !!currentStatus || !!currentOwnerType
  const rows = useMemo(() => groupPolicyReports(reports), [reports])

  const columns = useMemo<ColumnDef<IPolicyReportRow, unknown>[]>(
    () => [
      {
        id: 'subject',
        accessorFn: (row) => reportSubject(row.report),
        header: 'Component',
        cell: ({ row }) => (
          <PolicyReportPanel
            report={row.original.report}
            history={row.original.history}
            orgId={orgId}
            policyNameMap={policyNameMap}
            panelKey={`policy-report-${row.original.report?.id}`}
            triggerButton={{
              variant: 'ghost',
              className: panelTriggerClass,
              children: reportSubject(row.original.report),
            }}
          />
        ),
      },
      {
        id: 'source',
        accessorFn: (row) => row.report?.owner_type ?? '',
        header: 'Source',
        cell: ({ row }) => {
          const ownerType = row.original.report?.owner_type ?? ''
          const meta = OWNER_TYPE_LABELS[ownerType]
          return meta ? (
            <Badge size="sm" theme={meta.theme}>
              {meta.label}
            </Badge>
          ) : (
            <Text variant="subtext" theme="neutral">
              {ownerType || 'unknown'}
            </Text>
          )
        },
      },
      {
        id: 'status',
        accessorFn: (row) => row.report?.status?.status ?? '',
        header: 'Status',
        enableSorting: false,
        cell: ({ row }) => (
          <Status
            variant="badge"
            status={row.original.report?.status?.status ?? 'unknown'}
          />
        ),
      },
      {
        id: 'violations',
        accessorFn: (row) =>
          (row.report?.deny_count ?? 0) * 1000 + (row.report?.warn_count ?? 0),
        header: 'Violations',
        cell: ({ row }) => {
          const denyCount = row.original.report?.deny_count ?? 0
          const warnCount = row.original.report?.warn_count ?? 0

          if (!denyCount && !warnCount) {
            return (
              <Text variant="subtext" theme="neutral">
                None
              </Text>
            )
          }

          return (
            <span className="flex items-center gap-2">
              {denyCount ? (
                <Badge size="sm" theme="error">
                  {denyCount} denied
                </Badge>
              ) : null}
              {warnCount ? (
                <Badge size="sm" theme="warn">
                  {warnCount} warning{warnCount === 1 ? '' : 's'}
                </Badge>
              ) : null}
            </span>
          )
        },
      },
      {
        id: 'evaluated',
        accessorFn: (row) => getReportTime(row.report),
        header: 'Evaluated',
        cell: ({ row }) => (
          <Time
            variant="subtext"
            time={row.original.report?.evaluated_at ?? ''}
            format="relative"
          />
        ),
      },
      {
        id: 'evaluations',
        header: 'Evaluations',
        enableSorting: false,
        cell: ({ row }) => (
          <Text variant="subtext" theme="neutral" family="mono">
            {row.original.history.length + 1}
          </Text>
        ),
      },
    ],
    [orgId, policyNameMap]
  )

  return (
    <Table<IPolicyReportRow>
      columns={columns}
      data={rows}
      enableSearch={false}
      isLoading={isLoading}
      initialSorting={[{ id: 'evaluated', desc: true }]}
      emptyStateProps={{
        variant: 'policy',
        emptyTitle: hasActiveFilters ? 'No matching reports' : 'No evaluations yet',
        emptyMessage: hasActiveFilters
          ? 'No reports match the current filters.'
          : 'Evaluations appear here once a deploy or sandbox run triggers a policy check.',
      }}
    />
  )
}
