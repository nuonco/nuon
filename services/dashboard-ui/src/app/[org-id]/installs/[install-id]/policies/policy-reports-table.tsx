import { Banner } from '@/components/common/Banner'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import {
  PolicyReportsTable as Table,
  policyReportsTableColumns,
} from '@/components/policies/PolicyReportsTable'
import { getInstallPolicyReports } from '@/lib'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

export const PolicyReportsTable = async ({
  installId,
  orgId,
  status: statusFilter,
  ownerType,
}: {
  installId: string
  orgId: string
  status?: TPolicyReportStatus
  ownerType?: TPolicyReportOwnerType
}) => {
  const {
    data: reports,
    error,
    status,
  } = await getInstallPolicyReports({
    installId,
    orgId,
    status: statusFilter,
    ownerType,
  })

  return error && status !== 404 ? (
    <Banner theme="error">Can&apos;t load policy reports: {error?.error}</Banner>
  ) : (
    <Table
      reports={reports || []}
      orgId={orgId}
      installId={installId}
      currentStatus={statusFilter}
      currentOwnerType={ownerType}
    />
  )
}

export const PolicyReportsTableSkeleton = () => {
  return <TableSkeleton columns={policyReportsTableColumns} skeletonRows={5} />
}
