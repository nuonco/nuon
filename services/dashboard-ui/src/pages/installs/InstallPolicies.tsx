import { useParams, useSearchParams } from 'react-router-dom'
import { usePolling } from '@/hooks/use-polling'
import { BackToTop } from '@/components/common/BackToTop'
import { Banner } from '@/components/common/Banner'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import {
  PolicyReportsTable,
  policyReportsTableColumns,
} from '@/components/policies/PolicyReportsTable'
import { useOrg } from '@/hooks/use-org'
import { useInstall } from '@/hooks/use-install'
import type { TPolicyReport } from '@/types'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'

const PolicyReportsTableWrapper = ({
  installId,
  orgId,
  status,
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
    isLoading,
  } = usePolling<TPolicyReport[]>({
    path: `/api/ctl-api/v1/policy-reports?install_id=${installId}${status ? `&status=${status}` : ''}${ownerType ? `&owner_type=${ownerType}` : ''}`,
    shouldPoll: true,
    pollInterval: 30000,
  })

  if (error && error.status !== 404) {
    return (
      <Banner theme="error">
        Can&apos;t load policy reports: {error.message || 'Unknown error'}
      </Banner>
    )
  }

  if (isLoading && !reports) {
    return <TableSkeleton columns={policyReportsTableColumns} skeletonRows={5} />
  }

  return (
    <PolicyReportsTable
      reports={reports || []}
      orgId={orgId}
      installId={installId}
      currentStatus={status}
      currentOwnerType={ownerType}
    />
  )
}

export default function InstallPolicies() {
  const { installId, orgId } = useParams()
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const status = searchParams.get('status') as TPolicyReportStatus | undefined
  const ownerType = searchParams.get('owner_type') as
    | TPolicyReportOwnerType
    | undefined

  if (!installId || !orgId) {
    return null
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
          {
            path: `/${orgId}/installs/${installId}`,
            text: install?.name,
          },
          {
            path: `/${orgId}/installs/${installId}/policies`,
            text: 'Policies',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          Policy Evaluations
        </Text>
      </HeadingGroup>

      <div className="flex flex-auto">
        <PolicyReportsTableWrapper
          installId={installId}
          orgId={orgId}
          status={status}
          ownerType={ownerType}
        />
      </div>

      <BackToTop />
    </PageSection>
  )
}