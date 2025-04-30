import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallPageSubNav,
  Loading,
  RunnerMeta,
  RunnerHealthChart,
  RunnerPastJobs,
  RunnerUpcomingJobs,
  StatusBadge,
  Section,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { getInstall, getRunner } from '@/lib'

export default withPageAuthRequired(async function Runner({
  params,
  searchParams,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
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
            <div className="flex gap-6 items-start justify-start lg:gap-12 xl:gap-24 flex-wrap">
              <span className="flex flex-col gap-2">
                <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                  Status
                </Text>
                <StatusBadge
                  status={runner?.status}
                  description={runner?.status_description}
                  descriptionAlignment="left"
                  shouldPoll
                />
              </span>
              <RunnerMeta
                orgId={orgId}
                installId={installId}
                runnerId={runner.id}
              />
            </div>
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
                  offset={(searchParams['past-jobs'] as string) || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex-auto flex flex-col col-span-4">
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
                  offset={(searchParams['upcoming-jobs'] as string) || '0'}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
