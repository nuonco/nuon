import { Banner } from '@/components/common/Banner'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import {
  PolicyReportsTable as Table,
  policyReportsTableColumns,
} from '@/components/policies/PolicyReportsTable'
import { getInstallPolicyReports } from '@/lib'

export const PolicyReportsTable = async ({
  installId,
  orgId,
}: {
  installId: string
  orgId: string
}) => {
  const { data: reports, error, status } = await getInstallPolicyReports({
    installId,
    orgId,
  })

  return error && status !== 404 ? (
    <Banner theme="error">Can&apos;t load policy reports: {error?.error}</Banner>
  ) : (
    <Table reports={reports || []} orgId={orgId} installId={installId} />
  )
}

export const PolicyReportsTableSkeleton = () => {
  return <TableSkeleton columns={policyReportsTableColumns} skeletonRows={5} />
}
