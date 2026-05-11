import { useMemo } from 'react'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PolicyReportGroup } from '@/components/policies/PolicyReportGroup'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'
import type { TPolicyReport } from '@/types'

function getGroupKey(report: TPolicyReport): string {
  const ownerType = report.owner_type ?? 'unknown'
  if (report.component_id) {
    return `${ownerType}:${report.component_id}`
  }
  return ownerType
}

function getReportTime(report: TPolicyReport): number {
  return report.evaluated_at ? new Date(report.evaluated_at).getTime() : 0
}

interface IReportSection {
  reports: TPolicyReport[]
  orgId: string
  policyNameMap: Map<string, string>
}

function ReportSection({ reports, orgId, policyNameMap }: IReportSection) {
  if (reports.length === 0) return null

  const [latest, ...rest] = reports

  return (
    <div className="flex flex-col gap-2">
      <PolicyReportGroup
        report={latest}
        orgId={orgId}
        policyNameMap={policyNameMap}
      />

      {rest.length > 0 ? (
        <div className="flex flex-col gap-3 mt-8">
          <Text variant="h3" weight="strong">
            Older reports
          </Text>
          <Card className="!p-0 !gap-0 overflow-hidden">
            <div className="divide-y divide-cool-grey-200 dark:divide-dark-grey-600">
              {rest.map((report) => (
                <PolicyReportGroup
                  key={report.id}
                  report={report}
                  orgId={orgId}
                  policyNameMap={policyNameMap}
                  variant="embedded"
                />
              ))}
            </div>
          </Card>
        </div>
      ) : null}
    </div>
  )
}

export const PolicyReportsTable = ({
  reports,
  orgId,
  installId,
  policyNameMap,
  currentStatus,
  currentOwnerType,
}: {
  reports: TPolicyReport[]
  orgId: string
  installId: string
  policyNameMap: Map<string, string>
  currentStatus?: TPolicyReportStatus
  currentOwnerType?: TPolicyReportOwnerType
}) => {
  const hasActiveFilters = currentStatus || currentOwnerType

  const groupedReports = useMemo(() => {
    const groups = new Map<string, TPolicyReport[]>()
    for (const report of reports) {
      const key = getGroupKey(report)
      const arr = groups.get(key) ?? []
      arr.push(report)
      groups.set(key, arr)
    }

    const sections = Array.from(groups.entries()).map(([key, items]) => {
      const sorted = [...items].sort(
        (a, b) => getReportTime(b) - getReportTime(a)
      )
      return { key, reports: sorted }
    })

    sections.sort(
      (a, b) => getReportTime(b.reports[0]) - getReportTime(a.reports[0])
    )

    return sections
  }, [reports])

  if (reports.length === 0 && !hasActiveFilters) {
    return (
      <EmptyState
        className="py-12"
        variant="policy"
        emptyTitle="No evaluations yet"
        emptyMessage="Evaluations will appear here once a deploy or sandbox run triggers a policy check."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4 w-full">
      {reports.length === 0 ? (
        <EmptyState
          variant="policy"
          emptyTitle="No matching reports"
          emptyMessage="No reports match the current filters."
        />
      ) : (
        <div className="flex flex-col gap-6">
          {groupedReports.map((section) => (
            <ReportSection
              key={section.key}
              reports={section.reports}
              orgId={orgId}
              policyNameMap={policyNameMap}
            />
          ))}
        </div>
      )}
    </div>
  )
}
