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

  return (
    <div className="flex flex-col gap-4 w-full">
      <div className="flex items-center justify-between">
        <Text variant="subtext" theme="neutral">
          {reports.length} {reports.length === 1 ? 'report' : 'reports'}
        </Text>
        <PolicyReportsFilter
          currentStatus={currentStatus}
          currentOwnerType={currentOwnerType}
        />
      </div>

      {reports.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Text variant="body" weight="strong" theme="neutral">
            {hasActiveFilters ? 'No matching reports' : 'No policy evaluations'}
          </Text>
          <Text variant="subtext" theme="neutral" className="mt-1">
            {hasActiveFilters
              ? 'No reports match the current filters.'
              : 'Policy evaluations will appear here after deploys or sandbox runs.'}
          </Text>
        </div>
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
