// TODO(nnnat): remove once we have this API change on prod
// @ts-nocheck
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import { Warning } from '@phosphor-icons/react/dist/ssr'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  Config,
  ConfigContent,
  DashboardContent,
  Duration,
  ErrorFallback,
  ID,
  InstallCloudPlatform,
  InstallInputs,
  InstallInputsModal,
  InstallPageSubNav,
  InstallStatuses,
  Loading,
  StatusBadge,
  Section,
  SectionHeader,
  Text,
  Time,
  Markdown,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import {
  getInstall,
  getInstallCurrentInputs,
  getInstallReadme,
  getInstallRunnerGroup,
  getRunnerLatestHeartbeat,
} from '@/lib'
import { RUNNERS, USER_REPROVISION } from '@/utils'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const install = await getInstall({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
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
          runnerId={install?.runner_id}
        />
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <Section heading="README" className="md:col-span-8">
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={<Loading loadingText="Loading install README..." />}
            >
              <LoadInstallReadme installId={installId} orgId={orgId} />
            </Suspense>
          </ErrorBoundary>
        </Section>

        <div className="divide-y flex flex-col col-span-4">
          {RUNNERS ? (
            <Section className="flex-initial" heading="Runner">
              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={<Loading loadingText="Loading install runner..." />}
                >
                  <LoadRunnerGroup installId={installId} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </Section>
          ) : null}

          <Section className="flex-initial">
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={<Loading loadingText="Loading install inputs..." />}
              >
                <LoadInstallCurrentInputs installId={installId} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>

          <Section className="flex-initial" heading="Active sandbox">
            <div className="flex flex-col gap-8">
              <AppSandboxConfig sandboxConfig={install?.app_sandbox_config} />
              <AppSandboxVariables
                variables={install?.app_sandbox_config?.variables}
              />
            </div>
          </Section>

          <Section heading="Cloud platform">
            <InstallCloudPlatform install={install} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadInstallReadme: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const installReadme = await getInstallReadme({ installId, orgId }).catch(
    console.error
  )

  return installReadme ? (
    <div className="flex flex-col gap-3">
      {installReadme?.warnings?.length
        ? installReadme?.warnings?.map((warn, i) => (
            <span
              key={`${warn}-${i} `}
              className="flex items-center gap-3 w-full p-2 border rounded-md border-orange-400 bg-orange-300/20 text-orange-800 dark:border-orange-600 dark:bg-orange-600/5 dark:text-orange-600 text-base font-medium"
            >
              <Warning size={50} /> <span>{warn}</span>
            </span>
          ))
        : null}
      <Markdown content={installReadme?.readme} />
    </div>
  ) : (
    <Text variant="reg-12">No install README found</Text>
  )
}

const LoadInstallCurrentInputs: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const currentInputs = await getInstallCurrentInputs({ installId, orgId })

  return (
    <>
      <SectionHeader
        actions={
          currentInputs?.redacted_values ? (
            <InstallInputsModal currentInputs={currentInputs} />
          ) : undefined
        }
        className="mb-4"
        heading="Current inputs"
      />
      {currentInputs?.redacted_values ? (
        <InstallInputs currentInputs={currentInputs} />
      ) : (
        <Text>No inputs configured.</Text>
      )}
    </>
  )
}

const LoadRunnerGroup: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const runnerGroup = await getInstallRunnerGroup({ installId, orgId })

  return (
    <div className="flex flex-col gap-4">
      <div className="divide-y">
        {runnerGroup.runners?.length ? (
          runnerGroup.runners?.map((runner) => (
            <div key={runner?.id} className="flex flex-col gap-2">
              <span>
                <Text className="gap-3" variant="med-14">
                  <StatusBadge
                    status={runner?.status}
                    isStatusTextHidden
                    isWithoutBorder
                  />
                  <span>{runner?.display_name}</span>
                </Text>
                <ID id={runner?.id} />
              </span>

              <ErrorBoundary fallbackRender={ErrorFallback}>
                <Suspense
                  fallback={
                    <Loading loadingText="Loading runner heartbeat..." />
                  }
                >
                  <LoadRunnerHeartBeat runnerId={runner?.id} orgId={orgId} />
                </Suspense>
              </ErrorBoundary>
            </div>
          ))
        ) : (
          <Text>No runner found</Text>
        )}
      </div>
    </div>
  )
}

const LoadRunnerHeartBeat: FC<{ orgId: string; runnerId: string }> = async ({
  orgId,
  runnerId,
}) => {
  const heartbeat = await getRunnerLatestHeartbeat({ runnerId, orgId }).catch(
    console.error
  )

  return heartbeat ? (
    <Config>
      <ConfigContent label="Version" value={heartbeat?.version} />
      <ConfigContent
        label="Alive time"
        value={<Duration nanoseconds={heartbeat?.alive_time} />}
      />
      <ConfigContent
        label="Last seen"
        value={<Time time={heartbeat?.created_at} format="relative" />}
      />
    </Config>
  ) : (
    <Text>Runner not online.</Text>
  )
}
