import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { WorkflowTimelineSkeleton } from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import type { TPageProps } from '@/types'
import { getInstall, getOrg } from '@/lib'
import { Workflows, WorkflowsError } from './workflows'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstall({ installId, orgId })

  return {
    title: `Workflows | ${install.name} | Nuon`,
  }
}

export default async function InstallWorkflowsPage({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
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
            path: `/${orgId}/installs/${installId}/workflows`,
            text: 'Workflows',
          },
        ]}
      />
      <div className="flex items-center gap-4 justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Workflows
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <ShowDriftScan />
          <WorkflowTypeFilter />
        </div>
      </div>
      <ErrorBoundary fallback={<WorkflowsError />}>
        <Suspense fallback={<WorkflowTimelineSkeleton />}>
          <Workflows
            installId={installId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
            type={sp['type'] || ''}
            showDrift={sp['drifts'] !== 'false'}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  )
}
