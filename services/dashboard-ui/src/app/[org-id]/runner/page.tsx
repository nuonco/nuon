import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  ID,
  Loading,
  StatusBadge,
  RunnerHealthChart,
  RunnerHeartbeat,
  RunnerPastJobs,
  RunnerRecentJob,
  RunnerUpcomingJobs,
  Section,
  Text,
} from '@/components'
import { getOrg, getRunner } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Build runner`,
  }
}

export default withPageAuthRequired(async function OrgRunner({
  params,
  searchParams,
}) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const runnerId = org?.runner_group?.runners?.at(0)?.id
  const runner = await getRunner({ orgId, runnerId })

  if (org?.features?.['org-runner']) {
    return (
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}/runner`, text: 'Build runner' }]}
        heading={org?.name}
        headingUnderline={org?.id}
        statues={
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <StatusBadge
              status={org?.status}
              description={org?.status_description}
              descriptionAlignment="right"
              shouldPoll
            />
          </span>
        }
      >
        <div className="flex-auto md:grid md:grid-cols-12 divide-x">
          <div className="divide-y flex flex-col flex-auto col-span-8">
            <Section
              heading={
                <span>
                  <Text variant="med-14">{runner?.display_name} </Text>
                  <ID id={runner?.id} />
                </span>
              }
              className="flex-initial"
            >
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
                      runnerType="build"
                    />
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
          <div className="divide-y flex flex-col flex-auto col-span-4">
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
                  <RunnerRecentJob runnerId={runnerId} orgId={orgId} />
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
  } else {
    redirect(`/${orgId}/apps`)
  }
})
