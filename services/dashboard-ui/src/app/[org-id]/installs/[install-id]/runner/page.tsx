import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  DeprovisionRunnerModal,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  Link,
  Loading,
  Section,
  ShutdownRunnerModal,
  UpdateRunnerModal,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { ManageRunnerDropdown } from '@/components/Runners/ManageDropdown'
import { getInstallById, getRunnerById, getRunnerSettingsById } from '@/lib'
import { Activity } from './activity'
import { Details } from './details'
import { Health } from './health'
import { UpcomingJobs } from './upcoming-jobs'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Runner | ${install?.name} | Nuon`,
  }
}

export default async function Runner({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const { data: install } = await getInstallById({ installId, orgId })
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

  return (
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
              <Text isMuted>Managed By</Text>
              <Text>
                <FileCodeIcon />
                Config File
              </Text>
            </span>
          ) : null}
          <span className="flex flex-col gap-2">
            <Text isMuted>App config</Text>
            <Text>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
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
            <ErrorBoundary fallbackRender={ErrorFallback}>
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
            </ErrorBoundary>
          </Section>
          <Section className="flex-initial">
            <ErrorBoundary fallbackRender={ErrorFallback}>
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
            </ErrorBoundary>
          </Section>
          <Section heading="Completed jobs">
            <ErrorBoundary fallbackRender={ErrorFallback}>
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
            </ErrorBoundary>
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
            <ErrorBoundary fallbackRender={ErrorFallback}>
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
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}
