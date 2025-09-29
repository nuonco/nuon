import type { Metadata } from 'next'
import { redirect, notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  Loading,
  Notice,
  StatusBadge,
  ShutdownRunnerModal,
  UpdateRunnerModal,
  Section,
  Text,
} from '@/components'
import { ManageRunnerDropdown } from "@/components/Runners/ManageDropdown"
import { getRunnerById, getRunnerSettingsById, getOrgById } from '@/lib'
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

  if (org?.features?.['org-runner']) {
    return (
      <DashboardContent
        banner={
          runner?.status === 'error' ? (
            <Notice className="!border-none !rounded-none">
              Buld runner is unhealthy
            </Notice>
          ) : null
        }
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
                  <Health runnerId={runnerId} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
            <Section className="flex-initial">
              <ErrorBoundary fallbackRender={ErrorFallback}>
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
              <ManageRunnerDropdown runner={runner} settings={settings} />             
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
                  <UpcomingJobs
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
