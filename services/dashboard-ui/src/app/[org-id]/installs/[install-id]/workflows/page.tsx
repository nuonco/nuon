import type { Metadata } from 'next'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { WorkflowTimelineSkeleton } from '@/components/workflows/WorkflowTimeline'
import { ShowDriftScan } from '@/components/workflows/filters/ShowDriftScans'
import { WorkflowTypeFilter } from '@/components/workflows/filters/WorkflowTypeFilter'
import type { TPageProps } from '@/types'
import { getInstallById, getOrgById } from '@/lib'
import { Workflows, WorkflowsError } from './workflows'

// NOTE: old layout stuff
import {
  DashboardContent,
  Link as OldLink,
  Loading,
  InstallPageSubNav,
  InstallStatuses,
  InstallManagementDropdown,
  Section,
  Text as OldText,
  Time,
} from '@/components'
import { InstallWorkflows } from './install-workflows'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

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
    getInstallById({ installId, orgId }),
    getOrgById({ orgId }),
  ])

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
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
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows`,
          text: 'Workflows',
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      headingMeta={
        <>
          Last updated <Time time={install?.updated_at} format="relative" />
        </>
      }
      statues={
        <div className="flex items-start gap-8">
          {install?.metadata?.managed_by &&
          install?.metadata?.managed_by === 'nuon/cli/install-config' ? (
            <span className="flex flex-col gap-2">
              <OldText isMuted>Managed By</OldText>
              <OldText>
                <FileCodeIcon />
                Config File
              </OldText>
            </span>
          ) : null}
          <span className="flex flex-col gap-2">
            <OldText isMuted>App config</OldText>
            <OldText>
              <OldLink href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </OldLink>
            </OldText>
          </span>
          <InstallStatuses />

          <InstallManagementDropdown />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="Install workflows" className="overflow-auto">
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            <InstallWorkflows
              installId={installId}
              orgId={orgId}
              offset={sp['workflows'] || '0'}
            />
          </Suspense>
        </Section>
      </div>
    </DashboardContent>
  )
}
