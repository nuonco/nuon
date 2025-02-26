import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CaretRight, Heartbeat, Timer } from '@phosphor-icons/react/dist/ssr'
import {
  ClickToCopy,
  Config,
  ConfigContent,
  CancelRunnerJobButton,
  DashboardContent,
  Duration,
  EmptyStateGraphic,
  ErrorFallback,
  ID,
  InstallManagementDropdown,
  InstallStatuses,
  InstallPageSubNav,
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
  getAppLatestInputConfig,
  getInstall,
  getRunner,
  getRunnerJobs,
  getRunnerHealthChecks,
  getRunnerLatestHeartbeat,
} from '@/lib'
import { USER_REPROVISION, INSTALL_UPDATE } from '@/utils'

export default withPageAuthRequired(async function Runner({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const runnerId = params?.['runner-id'] as string
  const [install, runner, runnerHeartbeat] = await Promise.all([
    getInstall({ installId, orgId }),
    getRunner({
      orgId,
      runnerId,
    }),
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }).catch(console.error),
  ])

  const appInputConfigs =
    (await getAppLatestInputConfig({
      appId: install?.app_id,
      orgId,
    }).catch(console.error)) || undefined

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
              installId={installId}
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
              install={install}
              inputConfig={appInputConfigs}
              hasUpdateInstall={INSTALL_UPDATE}
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
                <LoadPastJobs
                  installId={installId}
                  runnerId={runnerId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>
        </div>
        <div className="divide-y flex-auto flex flex-col col-span-4">
          <Section className="flex-initial" heading="Recent job">
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
          </Section>
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
      </div>
    </DashboardContent>
  )
})

const LoadRunnerHeartbeat: FC<{
  orgId: string
  runnerId: string
}> = async ({ orgId, runnerId }) => {
  const healthChecks = await getRunnerHealthChecks({ orgId, runnerId })

  return <RunnerHeartbeatChart healthchecks={healthChecks} />
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
      groups: ['build', 'deploy', 'sync', 'actions'],
      statuses: ['finished', 'failed'],
    },
  })

  const job = runnerJobs?.[0]

  const isDeploy = job?.group === 'deploy' || job?.group === 'sync'
  const hrefPath = isDeploy
    ? `components/${job?.metadata?.install_component_id}/deploys/${job?.metadata?.deploy_id}`
    : `actions/${job?.metadata?.action_workflow_id}/${job?.metadata?.action_workflow_run_id}`

  return runnerJobs?.length ? (
    <div className="flex items-start justify-between">
      <Config>
        <ConfigContent
          label="Name"
          value={
            job?.metadata
              ? isDeploy
                ? job?.metadata?.component_name
                : job?.metadata?.action_workflow_name
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
        <Link
          className="text-sm"
          href={`/${orgId}/installs/${installId}/${hrefPath}`}
        >
          Details <CaretRight />
        </Link>
      ) : null}
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
      groups: ['build', 'deploy', 'sync', 'actions'],
      statuses: ['available', 'queued'],
    },
  })

  return (
    <>
      {runnerJobs?.length ? (
        <div className="divide-y flex-auto w-full">
          <div className="grid grid-cols-10 pb-1 mb-1">
            <Text
              className="col-span-4 font-normal leading-normal text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500"
              variant="med-12"
            >
              Name
            </Text>
            <Text
              className="col-span-4 font-normal leading-normal text-sm tracking-wide text-cool-grey-600 dark:text-cool-grey-500"
              variant="med-12"
            >
              Group
            </Text>
          </div>
          {runnerJobs?.map((job) => {
            const isDeploy = job?.group === 'deploy' || job?.group === 'sync'
            const jobType = isDeploy ? 'deploy' : 'workflow-run'

            return (
              <div
                className="flex items-center justify-between w-full py-2 grid grid-cols-10 gap-2"
                key={job.id}
              >
                <Text className="col-span-4">
                  {isDeploy
                    ? job?.metadata?.component_name
                    : job?.metadata?.action_workflow_name}
                </Text>

                <Text className="col-span-4">{runnerJobs?.[0]?.group}</Text>

                <div className="col-span-2">
                  <CancelRunnerJobButton
                    runnerJobId={job?.id}
                    orgId={orgId}
                    jobType={jobType}
                    variant="ghost"
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
  installId: string
  orgId: string
  runnerId: string
}> = async ({ installId, orgId, runnerId }) => {
  const runnerJobs = await getRunnerJobs({
    orgId,
    runnerId,
    options: {
      groups: ['build', 'deploy', 'sync', 'actions'],
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
                  <span>Unknown</span>
                )}{' '}
                /
                <span className="!inline truncate max-w-[100px]">
                  {job?.group}
                </span>
              </>
            ),
            time: job?.updated_at,
            href: job?.metadata
              ? `/${orgId}/installs/${installId}/${hrefPath}`
              : null,
            isMostRecent: i === 0,
          }
        })}
    />
  )
}
