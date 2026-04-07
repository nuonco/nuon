import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
import { PolicyReportGroup } from '@/components/policies/PolicyReportGroup'
import { PolicyReportsFilter } from '@/components/policies/PolicyReportsFilter'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'
import type { TPolicyReport } from '@/types'

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
      <div className="flex items-center justify-between">
        <Text variant="subtext" theme="neutral">
          {reports.length} {reports.length === 1 ? 'report' : 'reports'}
        </Text>
        <PolicyReportsFilter />
      </div>

      {reports.length === 0 ? (
        <EmptyState
          variant="policy"
          emptyTitle="No matching reports"
          emptyMessage="No reports match the current filters."
        />
      ) : (
        <div className="flex flex-col gap-3">
          {reports.map((report) => (
            <PolicyReportGroup
              key={report.id}
              report={report}
              orgId={orgId}
              policyNameMap={policyNameMap}
            />
          ))}
        </div>
      )}
    </div>
  )
}
