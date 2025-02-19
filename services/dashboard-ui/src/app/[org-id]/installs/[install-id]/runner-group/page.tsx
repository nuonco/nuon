import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { CaretRight, Heartbeat } from '@phosphor-icons/react/dist/ssr'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  CodeViewer,
  DashboardContent,
  Config,
  ConfigContent,
  Duration,
  InstallCloudPlatform,
  InstallManagementDropdown,
  InstallPageSubNav,
  InstallStatuses,
  Loading,
  StatusBadge,
  Expand,
  ErrorFallback,
  Link,
  Grid,
  Section,
  Time,
  Text,
  Truncate,
  ToolTip,
} from '@/components'
import {
  getAppLatestInputConfig,
  getInstall,
  getInstallRunnerGroup,
  getOrg,
  getRunner,
  getRunnerJobs,
  getRunnerLatestHeartbeat,
} from '@/lib'
import { USER_REPROVISION, INSTALL_UPDATE } from '@/utils'

export default withPageAuthRequired(async function RunnerGroup({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, runnerGroup, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallRunnerGroup({ installId, orgId }),
    getOrg({ orgId }),
  ])

  const appInputConfigs =
    (await getAppLatestInputConfig({
      appId: install?.app_id,
      orgId,
    }).catch(console.error)) || undefined

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
          text: install.name,
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
          runnerId={install?.runner_id}
        />
      }
    >
      <Section heading="Runners">
        {runnerGroup?.runners?.length ? (
          <Grid variant="3-cols">
            {runnerGroup?.runners?.map((runner) => (
              <ErrorBoundary key={runner?.id} fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <LoadingRunner display_name={runner?.display_name} />
                  }
                >
                  <LoadRunner
                    installId={installId}
                    runnerId={runner?.id}
                    orgId={orgId}
                  />
                </Suspense>
              </ErrorBoundary>
            ))}
          </Grid>
        ) : (
          <Text>No install runners</Text>
        )}
      </Section>
    </DashboardContent>
  )
})

const LoadingRunner: FC<{ display_name: string }> = ({ display_name }) => {
  return (
    <div className="border rounded px-3 py-2 flex items-center">
      <Loading loadingText={`${display_name} is loading...`} />
    </div>
  )
}

const LoadRunner: FC<{
  installId: string
  runnerId: string
  orgId: string
}> = async ({ installId, runnerId, orgId }) => {
  const [runner, runnerJobs, runnerHeartbeat] = await Promise.all([
    getRunner({
      orgId,
      runnerId,
    }),
    getRunnerJobs({
      orgId,
      runnerId,
      options: {
        limit: '4',
        groups: ['build', 'deploy', 'sync', 'actions'],
      },
    }),
    getRunnerLatestHeartbeat({
      orgId,
      runnerId,
    }).catch(console.error),
  ])

  return (
    <Expand
      parentClass="border rounded"
      headerClass="px-3 py-2"
      id={runner?.id}
      isOpen
      key={runner?.id}
      heading={
        <span className="flex items-center gap-3">
          <Text variant="med-14" className="gap-2">
            <span className="animate-pulse">
              <StatusBadge
                status={runner?.status}
                isStatusTextHidden
                isWithoutBorder
              />
            </span>
            <span>{runner?.display_name}</span>
          </Text>
          {runnerHeartbeat ? (
            <Text variant="reg-14">
              <Heartbeat className="animate-pulse" size={14} />
              <Time time={runnerHeartbeat?.created_at} format="relative" />
            </Text>
          ) : null}
        </span>
      }
      expandContent={
        <div className="flex flex-col border-t divide-y">
          <div className="flex justify-between items-start p-3">
            {runnerHeartbeat ? (
              <>
                <Config>
                  <ConfigContent
                    label="Version"
                    value={runnerHeartbeat?.version}
                  />
                  <ConfigContent
                    label="Alive time"
                    value={
                      <Duration nanoseconds={runnerHeartbeat?.alive_time} />
                    }
                  />
                </Config>
                <Text>
                  <Link
                    href={`/${orgId}/installs/${installId}/runner-group/${runnerId}`}
                  >
                    Details <CaretRight />
                  </Link>
                </Text>
              </>
            ) : (
              <Text>Runner is not online.</Text>
            )}
          </div>
          <div className="p-3">
            {runnerJobs?.length ? (
              <div className="divide-y">
                {runnerJobs?.map((job) => {
                  const name =
                    job?.group === 'deploy' || job?.group === 'sync'
                      ? job?.metadata?.component_name
                      : job?.metadata?.action_workflow_name
                  return (
                    <div
                      className="grid grid-cols-10 w-full py-3 first:pt-0 last:pb-0"
                      key={job?.id}
                    >
                      <Text className="col-span-3">
                        {name ? (
                          name?.length >= 12 ? (
                            <ToolTip tipContent={name} alignment="left">
                              <Truncate variant="small">{name}</Truncate>
                            </ToolTip>
                          ) : (
                            name
                          )
                        ) : (
                          'Unknown job'
                        )}
                      </Text>
                      <Time
                        className="col-span-4"
                        time={job?.updated_at}
                        format="relative"
                      />
                      <div className="col-span-3 flex justify-end">
                        <StatusBadge status={job?.status} />
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <Text variant="reg-12">No jobs to display.</Text>
            )}
          </div>
        </div>
      }
    />
  )
}
