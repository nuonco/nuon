import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { AsyncBoundary } from '@/components/common/AsyncBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCardSkeleton'
import { RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCardSkeleton'
import { RunnerRecentActivitySkeleton } from '@/components/runners/RunnerRecentActivitySkeleton'
import { ManagementDropdown } from '@/components/runners/management/ManagementDropdown'
import { getRunner, getRunnerSettings, getOrg } from '@/lib'
import { RunnerActivity, RunnerActivityError } from './runner-activity'
import { RunnerDetails, RunnerError } from './runner-details'
import { RunnerHealth, RunnerHealthError } from './runner-health'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrg({ orgId })

  return {
    title: `Builds | ${org.name} | Nuon`,
  }
}

export default async function OrgRunner({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const { data: org } = await getOrg({ orgId })
  const runnerId = org?.runner_group?.runners?.at(0)?.id
  const [{ data: runner, error }, { data: settings }] = await Promise.all([
    getRunner({
      orgId,
      runnerId,
    }),
    getRunnerSettings({
      orgId,
      runnerId,
    }),
  ])

  if (error) {
    notFound()
  }

  return (
    <PageLayout isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/runner`,
            text: 'Builds',
          },
        ]}
      />
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="strong" level={1}>
            Builds
          </Text>
          <Text theme="neutral">
            View your organizations build runner performance and activities.
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <TemporalLink namespace="runners" eventLoopId={runner?.id} />
          <ManagementDropdown runner={runner} settings={settings} />
        </div>
      </PageHeader>
      <PageContent>
        <PageSection className="flex-row gap-6">
          <AsyncBoundary
            errorFallback={<RunnerError />}
            loadingFallback={<RunnerDetailsCardSkeleton />}
          >
            <RunnerDetails org={org} />
          </AsyncBoundary>

          <AsyncBoundary
            errorFallback={<RunnerHealthError />}
            loadingFallback={<RunnerHealthCardSkeleton />}
          >
            <RunnerHealth org={org} />
          </AsyncBoundary>
        </PageSection>

        <div className="flex gap-6">
          <PageSection>
            <AsyncBoundary
              errorFallback={<RunnerActivityError />}
              loadingFallback={<RunnerRecentActivitySkeleton />}
            >
              <RunnerActivity org={org} offset={sp['offset'] || '0'} />
            </AsyncBoundary>
          </PageSection>
        </div>
      </PageContent>
    </PageLayout>
  )
}
