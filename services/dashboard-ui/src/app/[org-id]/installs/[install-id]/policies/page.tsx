import type { Metadata } from 'next'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getInstall, getOrg } from '@/lib'
import type {
  TPolicyReportOwnerType,
  TPolicyReportStatus,
} from '@/lib/ctl-api/installs/get-install-policy-reports'
import type { TPageProps } from '@/types'
import {
  PolicyReportsTable,
  PolicyReportsTableSkeleton,
} from './policy-reports-table'

type TInstallPoliciesPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPoliciesPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Policies | ${install?.name} | Nuon`,
  }
}

export default async function InstallPoliciesPage({
  params,
  searchParams,
}: TInstallPoliciesPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const resolvedSearchParams = await searchParams
  const status = resolvedSearchParams?.status as TPolicyReportStatus | undefined
  const ownerType = resolvedSearchParams?.owner_type as
    | TPolicyReportOwnerType
    | undefined

  const [{ data: install }, { data: org }] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

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
        <AsyncBoundary
          errorFallback={
            <Text variant="body" theme="neutral">
              Unable to load policy reports
            </Text>
          }
          loadingFallback={<PolicyReportsTableSkeleton />}
        >
          <PolicyReportsTable
            installId={installId}
            orgId={orgId}
            status={status}
            ownerType={ownerType}
          />
        </AsyncBoundary>
      </div>
    </PageSection>
  )
}
