import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageSection } from '@/components/layout/PageSection'
import { RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCardSkeleton'
import { RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCardSkeleton'
import { RunnerRecentActivitySkeleton } from '@/components/runners/RunnerRecentActivitySkeleton'
import { getRunnerById, getRunnerSettingsById, getOrgById } from '@/lib'
import { RunnerActivity, RunnerActivityError } from './runner-activity'
import { RunnerDetails, RunnerError } from './runner-details'
import { RunnerHealth, RunnerHealthError } from './runner-health'

// NOTE: old layout components
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  Notice,
  StatusBadge,
  Section,
  Text as OldText,
} from '@/components'
import { ManageRunnerDropdown } from '@/components/old/OldRunners/ManageDropdown'
import { Activity } from './activity'
import { Details } from './details'
import { Health } from './health'
import { UpcomingJobs } from './upcoming-jobs'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Build runner | ${org.name} | Nuon`,
  }
}

export default async function OrgRunner({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const { data: org } = await getOrgById({ orgId })
  const runnerId = org?.runner_group?.runners?.at(0)?.id
  const [{ data: runner, error }, { data: settings }] = await Promise.all([
    getRunnerById({
      orgId,
      runnerId,
    }),
    getRunnerSettingsById({
      orgId,
      runnerId,
    }),
  ])

  if (error) {
    notFound()
  }

  return org?.features?.['stratus-layout'] ? (
    <PageLayout
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/runner`,
            text: 'Runner',
          },
        ],
      }}
      isScrollable
    >
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="strong" level={1}>
            Build runner
          </Text>
          <Text theme="neutral">
            View your organizations build runner performance and activities.
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          <TemporalLink namespace="runners" eventLoopId={runner?.id} />
          <ManageRunnerDropdown runner={runner} settings={settings} />
        </div>
      </PageHeader>
      <PageContent>
        <PageSection className="flex-row gap-6">
          <ErrorBoundary fallback={<RunnerError />}>
            <Suspense
              fallback={<RunnerDetailsCardSkeleton className="flex-initial" />}
            >
              <RunnerDetails org={org} />
            </Suspense>
          </ErrorBoundary>
          <ErrorBoundary fallback={<RunnerHealthError />}>
            <Suspense
              fallback={<RunnerHealthCardSkeleton className="flex-auto" />}
            >
              <RunnerHealth org={org} />
            </Suspense>
          </ErrorBoundary>
        </PageSection>

        <div className="flex gap-6">
          <PageSection>
            <ErrorBoundary fallback={<RunnerActivityError />}>
              <Suspense fallback={<RunnerRecentActivitySkeleton />}>
                <RunnerActivity org={org} offset={sp['offset'] || '0'} />
              </Suspense>
            </ErrorBoundary>
          </PageSection>
        </div>
      </PageContent>
    </PageLayout>
  ) : (
    <DashboardContent
      banner={
        runner?.status === 'error' ? (
          <Notice className="!border-none !rounded-none">
            Build runner is unhealthy
          </Notice>
        ) : null
      }
      breadcrumb={[{ href: `/${orgId}/runner`, text: 'Build runner' }]}
      heading={org?.name}
      headingUnderline={org?.id}
      statues={
        <span className="flex flex-col gap-2">
          <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
            Status
          </OldText>
          <StatusBadge
            status={org?.status}
            description={org?.status_description}
            descriptionAlignment="right"
          />
        </span>
      }
    >
      <div className="flex-auto md:grid md:grid-cols-12 divide-x">
        <div className="divide-y flex flex-col flex-auto col-span-8">
          <Section className="flex-initial" heading="Health">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner health status..."
                  />
                }
              >
                <Health runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </OldErrorBoundary>
          </Section>
          <Section className="flex-initial">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner details..."
                  />
                }
              >
                <Details orgId={orgId} runner={runner} settings={settings} />
              </Suspense>
            </OldErrorBoundary>
          </Section>
          <Section heading="Completed jobs">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner jobs..."
                  />
                }
              >
                <Activity
                  runnerId={runnerId}
                  orgId={orgId}
                  offset={(sp['past-jobs'] as string) || '0'}
                />
              </Suspense>
            </OldErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex flex-col flex-auto col-span-4">
          <Section heading="Runner controls" className="flex-initial">
            <ManageRunnerDropdown runner={runner} settings={settings} />
          </Section>
          <Section className="flex-initial" heading="Upcoming jobs ">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading upcoming jobs..."
                  />
                }
              >
                <UpcomingJobs
                  runnerId={runnerId}
                  orgId={orgId}
                  offset={(sp['upcoming-jobs'] as string) || '0'}
                />
              </Suspense>
            </OldErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
