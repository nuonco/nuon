import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  DeprovisionRunnerModal,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  Link,
  Loading,
  RunnerMeta,
  RunnerHealthChart,
  RunnerPastJobs,
  RunnerUpcomingJobs,
  Section,
  ShutdownRunnerModal,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { getInstall, getRunner } from '@/lib'

export default async function Runner({
  params,
  searchParams,
}) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const install = await getInstall({ installId, orgId })
  const runner = await getRunner({
    orgId,
    runnerId: install.runner_id,
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
          <span className="flex flex-col gap-2">
            <Text isMuted>App config</Text>
            <Text>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </span>
          <InstallStatuses initInstall={install} shouldPoll />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
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
                <RunnerHealthChart runnerId={runner.id} orgId={orgId} />
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
                              
                  <RunnerMeta
                    orgId={orgId}
                    installId={installId}
                    runner={runner}
                  />

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
                <RunnerPastJobs
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
            <div className="flex items-center gap-4">
              <ShutdownRunnerModal orgId={orgId} runnerId={runner?.id} />
              <DeprovisionRunnerModal />
            </div>
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
                <RunnerUpcomingJobs
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
