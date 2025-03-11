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
import { getInstall, getRunner } from '@/lib'
import { USER_REPROVISION } from '@/utils'

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
          text: runner?.display_name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={
        <div className="flex items-start gap-8">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Created
            </Text>
            <Time variant="reg-12" time={install?.created_at} />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Updated
            </Text>
            <Time variant="reg-12" time={install?.updated_at} />
          </span>
          <InstallStatuses initInstall={install} shouldPoll />
          {USER_REPROVISION ? (
            <InstallManagementDropdown
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
              install={install}
            />
          ) : null}
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
        <div className="divide-y flex flex-col flex-auto col-span-8">
          <Section
            className="flex-initial"
            heading={
              <span>
                <Text variant="med-14">{runner?.display_name} </Text>
                <ID id={runner?.id} />
              </span>
            }
          >
            <div className="flex gap-6 items-start justify-start">
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
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading loadingText="Loading runner heartbeat..." />
                  }
                >
                  <RunnerHeartbeat runnerId={runnerId} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </div>
          </Section>
          <Section className="flex-initial" heading="Health status">
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
          <Section heading="Job run history">
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
        <div className="divide-y flex-auto flex flex-col col-span-4">
          <Section className="flex-initial" heading="Recent job">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    variant="stack"
                    loadingText="Loading recent job..."
                  />
                }
              >
                <RunnerRecentJob
                  runnerId={runnerId}
                  orgId={orgId}
                  groups={['deploy', 'sync', 'actions']}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
          <Section className="flex-initial" heading="Upcoming jobs ">
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
                  runnerId={runnerId}
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
