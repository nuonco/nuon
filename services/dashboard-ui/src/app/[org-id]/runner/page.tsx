import type { Metadata } from 'next'
import { redirect } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  StatusBadge,
  RunnerHealthChart,
  RunnerMeta,
  RunnerPastJobs,
  RunnerUpcomingJobs,
  ShutdownRunnerModal,
  Section,
  Text,
} from '@/components'
import { getRunner, getOrg } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Build runner`,
  }
}

export default async function OrgRunner({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const org = await getOrg({ orgId })
  const runnerId = org?.runner_group?.runners?.at(0)?.id
  const [runner] = await Promise.all([
    getRunner({
      orgId,
      runnerId,
    }),
  ])

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
                  <RunnerHealthChart runnerId={runnerId} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
            <Section className="flex-initial">
              <RunnerMeta orgId={orgId} runner={runner} />
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
                    runnerId={runnerId}
                    orgId={orgId}
                    offset={(sp['past-jobs'] as string) || '0'}
                  />
                </Suspense>
              </ErrorBoundary>
            </Section>
          </div>
          <div className="divide-y flex flex-col flex-auto col-span-4">
            <Section heading="Runner controls" className="flex-initial">
              <ShutdownRunnerModal orgId={orgId} runnerId={runner?.id} />
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
                    offset={(sp['upcoming-jobs'] as string) || '0'}
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
}
