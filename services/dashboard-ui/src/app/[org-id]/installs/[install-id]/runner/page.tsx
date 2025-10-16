import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { PageSection } from '@/components/layout/PageSection'
import { RunnerRecentActivitySkeleton } from '@/components/runners/RunnerRecentActivitySkeleton'
import { RunnerDetailsCardSkeleton } from '@/components/runners/RunnerDetailsCardSkeleton'
import { RunnerHealthCardSkeleton } from '@/components/runners/RunnerHealthCardSkeleton'
import { Text } from '@/components/common/Text'
import {
  getInstallById,
  getRunnerById,
  getRunnerSettingsById,
  getOrgById,
} from '@/lib'
import { TPageProps } from '@/types'
import { RunnerActivity, RunnerActivityError } from './runner-activity'
import { RunnerDetails, RunnerDetailsError } from './runner-details'
import { RunnerHealth, RunnerHealthError } from './runner-health'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  Link as OldLink,
  Loading,
  Section,
  Text as OldText,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { ManageRunnerDropdown } from '@/components/OldRunners/ManageDropdown'
import { Activity } from './activity'
import { Details } from './details'
import { Health } from './health'
import { UpcomingJobs } from './upcoming-jobs'

type TInstallPageProps = TPageProps<'org-id' | 'install-id'>

export async function generateMetadata({
  params,
}: TInstallPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Runner | ${install?.name} | Nuon`,
  }
}

export default async function Runner({
  params,
  searchParams,
}: TInstallPageProps) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const [{ data: install }, { data: org }] = await Promise.all([
    getInstallById({ installId, orgId }),
    getOrgById({ orgId }),
  ])
  const [{ data: runner, error }, { data: settings }] = await Promise.all([
    getRunnerById({
      orgId,
      runnerId: install.runner_id,
    }),
    getRunnerSettingsById({
      orgId,
      runnerId: install.runner_id,
    }),
  ])

  if (error) {
    notFound()
  }

  return org?.features?.['stratus-layout'] ? (
    <PageSection className="@container" isScrollable>
      <div className="flex gap-4 justify-between">
        <hgroup>
          <Text variant="base" weight="strong">
            Install runner
          </Text>
        </hgroup>
        <ManageRunnerDropdown
          runner={runner}
          settings={settings}
          isInstallRunner
        />
      </div>

      <div className="flex flex-col @min-4xl:flex-row gap-6">
        <ErrorBoundary fallback={<RunnerDetailsError />}>
          <Suspense
            fallback={<RunnerDetailsCardSkeleton className="flex-initial" />}
          >
            <RunnerDetails orgId={orgId} runnerId={install?.runner_id} settings={settings} />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary fallback={<RunnerHealthError />}>
          <Suspense
            fallback={<RunnerHealthCardSkeleton className="flex-auto" />}
          >
            <RunnerHealth orgId={orgId} runnerId={install.runner_id} />
          </Suspense>
        </ErrorBoundary>
      </div>

      <div className="flex flex-col gap-6">
        <ErrorBoundary fallback={<RunnerActivityError />}>
          <Suspense fallback={<RunnerRecentActivitySkeleton />}>
            <RunnerActivity
              orgId={orgId}
              offset={sp['offset'] || '0'}
              runnerId={install.runner_id}
            />
          </Suspense>
        </ErrorBoundary>
      </div>
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
          href: `/${orgId}/installs/${install.id}/runner`,
          text: 'Runner',
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
                <Health runnerId={runner.id} orgId={orgId} />
              </Suspense>
            </OldErrorBoundary>
          </Section>
          <Section className="flex-initial">
            <OldErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner metadata..."
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
                  runnerId={runner.id}
                  orgId={orgId}
                  offset={(sp['past-jobs'] as string) || '0'}
                />
              </Suspense>
            </OldErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex-auto flex flex-col col-span-4">
          <Section heading="Runner controls" className="flex-initial">
            <ManageRunnerDropdown
              runner={runner}
              settings={settings}
              isInstallRunner
            />
          </Section>
          <Section heading="Upcoming jobs ">
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
                  runnerId={runner.id}
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
