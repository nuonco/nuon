import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { CaretRight, Heartbeat, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  Config,
  ConfigContent,
  CancelRunnerJobButton,
  DashboardContent,
  Duration,
  EmptyStateGraphic,
  ErrorFallback,
  ID,
  Link,
  Loading,
  StatusBadge,
  RunnerHeartbeatChart,
  Section,
  Text,
  Time,
  Timeline,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getOrg,
  getRunner,
  getRunnerHealthChecks,
  getRunnerJobs,
  getRunnerLatestHeartbeat,
} from '@/lib'

// TODO(nnnat): future org level dashboard
export default async function OrgDashboard({ params }) {
  const orgId = params?.['org-id'] as string
  const org = await getOrg({ orgId })
  const runnerId = org?.runner_group?.runners?.at(0)?.id
  const [runner, runnerHeartbeat] = await Promise.all([
    getRunner({ orgId, runnerId }),
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }).catch(console.error),
  ])

  return (
    <DashboardContent
      breadcrumb={[{ href: `/${orgId}`, text: org?.name }]}
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
              {runnerHeartbeat ? (
                <>
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Version
                    </Text>
                    <Text variant="med-12">{runnerHeartbeat?.version}</Text>
                  </span>
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Alive time
                    </Text>
                    <Text>
                      <Timer size={14} />
                      <Duration
                        nanoseconds={runnerHeartbeat.alive_time}
                        variant="med-12"
                      />
                    </Text>
                  </span>
                  <span className="flex flex-col gap-2">
                    <Text className="text-cool-grey-600 dark:text-cool-grey-500">
                      Last heartbeat seen
                    </Text>
                    <Text>
                      <Heartbeat size={14} />
                      <Time
                        time={runnerHeartbeat.created_at}
                        format="relative"
                        variant="med-12"
                      />
                    </Text>
                  </span>
                </>
              ) : null}
            </div>
          </Section>
          <Section className="flex-initial" heading="Health status">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading loadingText="Loading runner health status..." />
                }
              >
                <LoadRunnerHeartbeat runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
          <Section heading="Job run history">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading runner jobs..." />}
              >
                <LoadPastJobs runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex flex-col flex-auto col-span-4">
          <Section className="flex-initial" heading="Recent job">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading recent job..." />}
              >
                <LoadRecentJob runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
          <Section className="flex-initial" heading="Upcoming jobs ">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading upcoming jobs..." />}
              >
                <LoadUpcomingJobs runnerId={runnerId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
}

const LoadRunnerHeartbeat: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const healthChecks = await getRunnerHealthChecks({ orgId, runnerId })

  return <RunnerHeartbeatChart healthchecks={healthChecks} />
}

const LoadRecentJob: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      limit: '1',
      groups: ['build', 'sandbox'],
      statuses: ['finished', 'failed'],
    },
  })

  const job = runnerJobs?.[0]

  const isBuild = job?.group === 'build'
  const hrefPath = isBuild
    ? `apps/${job?.metadata?.app_id}/components/${job?.metadata?.component_id}/builds/${job?.metadata?.component_build_id}`
    : `installs/${job?.metadata?.install_id}/runs/${job?.metadata?.sandbox_run_id}`

  return runnerJobs?.length ? (
    <div className="flex items-start justify-between">
      <Config>
        <ConfigContent
          label="Name"
          value={
            job?.metadata
              ? isBuild
                ? job?.metadata?.component_name
                : job?.metadata?.sandbox_run_type
              : 'Unknown'
          }
        />

        <ConfigContent label="Group" value={job?.group} />

        <ConfigContent
          label="Status"
          value={
            <span className="flex items-center gap-2">
              <StatusBadge
                status={job?.status}
                isWithoutBorder
                isStatusTextHidden
              />
              {job?.status}
            </span>
          }
        />
      </Config>
      {job?.metadata ? (
        <Link className="text-sm" href={`/${orgId}/${hrefPath}`}>
          Details <CaretRight />
        </Link>
      ) : null}
    </div>
  ) : (
    <Text>No job to show.</Text>
  )
}

const LoadUpcomingJobs: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      groups: ['build', 'sandbox'],
      statuses: ['available', 'queued'],
    },
  })

  return (
    <>
      {runnerJobs?.length ? (
        <div className="divide-y flex-auto w-full">
          {runnerJobs?.map((job) => {
            const isBuild = job?.group === 'build'
            const jobType = isBuild ? 'build' : 'sandbox-run'

            return (
              <div
                className="flex items-center justify-between w-full py-3"
                key={job.id}
              >
                <Config>
                  <ConfigContent
                    label="Name"
                    value={
                      isBuild
                        ? job?.metadata?.component_name
                        : job?.metadata?.sandbox_run_type
                    }
                  />

                  <ConfigContent label="Group" value={runnerJobs?.[0]?.group} />
                </Config>
                <div className="">
                  <CancelRunnerJobButton
                    runnerJobId={job?.id}
                    orgId={orgId}
                    jobType={jobType}
                  />
                </div>
              </div>
            )
          })}
        </div>
      ) : (
        <div className="m-auto flex flex-col items-center max-w-[200px] my-6">
          <EmptyStateGraphic />
          <Text className="mt-6" variant="med-14">
            No upcoming jobs
          </Text>
          <Text variant="reg-12" className="text-center">
            Runner jobs will appear here as they become available and queued.
          </Text>
        </div>
      )}
    </>
  )
}

const LoadPastJobs: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      groups: ['build', 'sandbox'],
      limit: '10',
    },
  })

  return (
    <Timeline
      emptyMessage="No runner jobs have happened yet."
      events={runnerJobs
        ?.filter((job) => job?.group !== 'operations')
        .map((job, i) => {
          const isBuild = job?.group === 'build'
          const hrefPath = isBuild
            ? `apps/${job?.metadata?.app_id}/components/${job?.metadata?.component_id}/builds/${job?.metadata?.component_build_id}`
            : `installs/${job?.metadata?.install_id}/runs/${job?.metadata?.sandbox_run_id}`
          const name = isBuild
            ? job?.metadata?.component_name
            : job?.metadata?.sandbox_run_type

          return {
            id: job?.id,
            status: job?.status,
            underline: (
              <>
                {name ? (
                  name?.length >= 12 ? (
                    <ToolTip tipContent={name} alignment="right">
                      <Truncate variant="small">{name}</Truncate>
                    </ToolTip>
                  ) : (
                    name
                  )
                ) : (
                  <span>Unknown</span>
                )}{' '}
                /
                <span className="!inline truncate max-w-[100px]">
                  {job?.group}
                </span>
              </>
            ),
            time: job?.updated_at,
            href: job?.metadata ? `/${orgId}/${hrefPath}` : null,
            isMostRecent: i === 0,
          }
        })}
    />
  )
}
