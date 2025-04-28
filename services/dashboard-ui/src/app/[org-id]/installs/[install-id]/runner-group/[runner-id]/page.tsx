import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  ID,
  InstallStatuses,
  InstallPageSubNav,
  Loading,
  RunnerHealthChart,
  RunnerHeartbeat,
  RunnerPastJobs,
  RunnerRecentJob,
  RunnerUpcomingJobs,
  StatusBadge,
  Section,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { getInstall, getRunner, getAppLatestRunnerConfig } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const runnerId = params?.['runner-id'] as string
  const [install, runner] = await Promise.all([
    getInstall({ installId, orgId }),
    getRunner({ runnerId, orgId }),
  ])

  return {
    title: `${install.name} | ${runner.display_name}`,
  }
}

export default withPageAuthRequired(async function Runner({
  params,
  searchParams,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const runnerId = params?.['runner-id'] as string
  const [install, runner] = await Promise.all([
    getInstall({ installId, orgId }),
    getRunner({
      orgId,
      runnerId,
    }),
  ])

  const appRunnerConfig = await getAppLatestRunnerConfig({
    orgId,
    appId: install.app_id,
  })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/runner-group/${runnerId}`,
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
          <InstallStatuses initInstall={install} shouldPoll />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
        </div>
      }
      meta={
        <InstallPageSubNav
          installId={installId}
          orgId={orgId}
          runnerId={runnerId}
        />
      }
    >
      <div className="flex-auto md:grid md:grid-cols-12 divide-x">
        <div className="divide-y flex flex-col flex-auto col-span-7">
          <Section heading="Jobs">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner jobs..."
                  />
                }
              >
                <RunnerPastJobs
                  runnerId={runnerId}
                  orgId={orgId}
                  offset={(searchParams['past-jobs'] as string) || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex-auto flex flex-col col-span-5">
          <Section className="flex-initial" heading="Status">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading runner health status..."
                  />
                }
              >
                <RunnerHealthChart runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
          <Section className="flex-initial">
            <div className="grid gap-6 lg:grid-cols-2">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <span className="flex self-end">
                      <Loading loadingText="Loading runner heartbeat..." />
                    </span>
                  }
                >
                  <RunnerHeartbeat
                    runnerId={runnerId}
                    orgId={orgId}
                    runnerType={appRunnerConfig.app_runner_type}
                  />
                </Suspense>
              </ErrorBoundary>
            </div>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
