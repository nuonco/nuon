import { useSearchParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { BackToTop } from '@/components/common/BackToTop'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { TableSkeleton } from '@/components/common/TableSkeleton'
import { Text } from '@/components/common/Text'
import { PolicyReportsTable, policyReportsTableColumns } from '@/components/policies/PolicyReportsTable'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallPolicyReports } from '@/lib'
import type { TPolicyReportOwnerType, TPolicyReportStatus } from '@/lib/ctl-api/installs/get-install-policy-reports'

const CONTAINER_ID = 'install-policies-page'

export const Policies = () => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams] = useSearchParams()

  const status = searchParams.get('status') as TPolicyReportStatus | null
  const ownerType = searchParams.get('owner_type') as TPolicyReportOwnerType | null

  const { data: reportsResult, isLoading } = useQuery({
    queryKey: ['install-policy-reports', org?.id, install?.id, status, ownerType],
    queryFn: () =>
      getInstallPolicyReports({
        orgId: org.id,
        installId: install.id,
        status: status || undefined,
        ownerType: ownerType || undefined,
      }),
    enabled: !!org?.id && !!install?.id,
  })

  const reports = reportsResult ?? []

  return (
    <PageSection id={CONTAINER_ID} isScrollable>
      <PageTitle title={`Policies | ${install?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/installs`, text: 'Installs' },
          { path: `/${org?.id}/installs/${install?.id}`, text: install?.name },
          {
            path: `/${org?.id}/installs/${install?.id}/policies`,
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
        {isLoading ? (
          <TableSkeleton columns={policyReportsTableColumns} skeletonRows={5} />
        ) : (
          <PolicyReportsTable
            reports={reports}
            orgId={org?.id ?? ''}
            installId={install?.id ?? ''}
            currentStatus={status || undefined}
            currentOwnerType={ownerType || undefined}
          />
        )}
      </div>
      <BackToTop containerId={CONTAINER_ID} />
    </PageSection>
  )
}
