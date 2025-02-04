import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CaretRight, Heartbeat, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  DashboardContent,
  Config,
  ConfigContent,
  CancelRunnerJobButton,
  Duration,
  ErrorFallback,
  Loading,
  StatusBadge,
  Link,
  Section,
  Text,
  Time,
  Timeline,
  ToolTip,
  ClickToCopy,
  Truncate,
} from '@/components'
import {
  getInstall,
  getOrg,
  getRunner,
  getRunnerJobs,
  getRunnerHeartbeat,
  getRunnerLatestHeartbeat,
} from '@/lib'

import { RunnerHeartbeatChart } from '@/components/RunnerHeartbeatChart'

export default withPageAuthRequired(async function Runner({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const runnerId = params?.['runner-id'] as string
  const [install, runner, org, runnerHeartbeat] = await Promise.all([
    getInstall({ installId, orgId }),
    getRunner({
      orgId,
      runnerId,
    }),
    getOrg({ orgId }),
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}/runner-group`,
          text: install.name,
        },
      ]}
      heading={runner.display_name}
      headingUnderline={runner.id}
      statues={
        <div className="flex gap-6 items-start justify-start">
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
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Status
            </Text>
            <StatusBadge status={runner?.status} shouldPoll />
          </span>
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Install
            </Text>
            <Text variant="med-12">{install.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={install.id}>
                <ClickToCopy>
                  <Truncate variant="small">{install.id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>
        </div>
      }
    >
      <div className="flex flex-col md:flex-row flex-auto divide-x">
        <div className="divide-y flex flex-col flex-auto">
          <Section className="flex-initial" heading={`Heartbeat`}>
            <Text variant="reg-12">TBD</Text>

            {/* <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                fallback={<Loading loadingText="Loading runner heartbeat..." />}
                >
                <LoadRunnerHeartbeat runnerId={runnerId} orgId={orgId} />
                </Suspense>
                </ErrorBoundary> */}
          </Section>
          {/* <Section className="flex-initial" heading="Recent job">
              <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
              fallback={<Loading loadingText="Loading recent job..." />}
              >
              <LoadRecentJob
              installId={installId}
              runnerId={runnerId}
              orgId={orgId}
              />
              </Suspense>
              </ErrorBoundary>
              </Section> */}

          <Section className="flex-initial" heading="Upcoming jobs ">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading upcoming jobs..." />}
              >
                <LoadUpcomingJobs
                  installId={installId}
                  runnerId={runnerId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex-auto flex flex-col lg:min-w-[450px] lg:max-w-[450px]">
          <Section heading="Past job runs">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading runner jobs..." />}
              >
                <LoadPastJobs
                  installId={installId}
                  runnerId={runnerId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadRunnerHeartbeat: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const heartBeats = await getRunnerHeartbeat({ orgId, runnerId })

  return <RunnerHeartbeatChart />
}

const LoadRecentJob: FC<{
  installId: string
  orgId: string
  runnerId: string
}> = async ({ installId, orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      limit: '1',
    },
  })

  const isDeploy =
    runnerJobs?.[0]?.group === 'deploy' || runnerJobs?.[0]?.group === 'sync'
  const hrefPath = isDeploy
    ? `components/${runnerJobs?.[0]?.metadata?.install_component_id}/deploys/${runnerJobs?.[0]?.metadata?.deploy_id}`
    : `actions/${runnerJobs?.[0]?.metadata?.action_workflow_id}/${runnerJobs?.[0]?.metadata?.action_workflow_run_id}`

  return runnerJobs?.length ? (
    <div className="flex items-start justify-between">
      <Config>
        <ConfigContent
          label="Name"
          value={
            isDeploy
              ? runnerJobs?.[0]?.metadata?.component_name
              : runnerJobs?.[0]?.metadata?.action_workflow_name
          }
        />

        <ConfigContent label="Group" value={runnerJobs?.[0]?.group} />

        <ConfigContent label="Type" value={runnerJobs?.[0]?.type} />
      </Config>
      <Link
        className="text-sm"
        href={`/${orgId}/installs/${installId}/${hrefPath}`}
      >
        Details <CaretRight />
      </Link>
    </div>
  ) : (
    <Text>No job to show.</Text>
  )
}

const LoadUpcomingJobs: FC<{
  installId: string
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      statuses: ['available', 'queued'],
    },
  })

  return (
    <>
      {runnerJobs?.length ? (
        <div className="divide-y">
          {runnerJobs?.map((job) => {
            const isDeploy = job?.group === 'deploy' || job?.group === 'sync'
            const jobType = isDeploy ? 'deploy' : 'workflow-run'

            return (
              <div
                className="flex items-center justify-between w-full py-3"
                key={job.id}
              >
                <Config>
                  <ConfigContent
                    label="Name"
                    value={
                      isDeploy
                        ? job?.metadata?.component_name
                        : job?.metadata?.action_workflow_name
                    }
                  />

                  <ConfigContent label="Group" value={runnerJobs?.[0]?.group} />

                  <ConfigContent label="Type" value={runnerJobs?.[0]?.type} />
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
        <Text>No upcoming job.</Text>
      )}
    </>
  )
}

const LoadPastJobs: FC<{
  installId: string
  orgId: string
  runnerId: string
}> = async ({ installId, orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      limit: '10',
    },
  })

  return (
    <Timeline
      emptyMessage="No runner jobs have happened yet."
      events={runnerJobs
        ?.filter((job) => job?.group !== 'operations')
        .map((job, i) => {
          const hrefPath =
            job?.group === 'deploy' || job?.group === 'sync'
              ? `components/${job?.metadata?.install_component_id}/deploys/${job?.metadata?.deploy_id}`
              : `actions/${job?.metadata?.action_workflow_id}/${job?.metadata?.action_workflow_run_id}`
          const name =
            job?.group === 'deploy' || job?.group === 'sync'
              ? job?.metadata?.component_name
              : job?.metadata?.action_workflow_name

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
                  <span>Not attempted</span>
                )}{' '}
                /
                <span className="!inline truncate max-w-[100px]">
                  {job?.group}
                </span>
              </>
            ),
            time: job?.updated_at,
            href: `/${orgId}/installs/${installId}/${hrefPath}`,
            isMostRecent: i === 0,
          }
        })}
    />
  )
}
