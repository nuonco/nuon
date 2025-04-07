import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  CalendarBlank,
  CaretRight,
  Timer,
} from '@phosphor-icons/react/dist/ssr'
import {
  CancelRunnerJobButton,
  ClickToCopy,
  ClickToCopyButton,
  CodeViewer,
  ComponentConfiguration,
  DashboardContent,
  DeployStatus,
  Duration,
  ErrorFallback,
  InstallDeployIntermediateData,
  Link,
  Loading,
  LogStreamProvider,
  OperationLogsSection,
  StatusBadge,
  Section,
  Text,
  Time,
  ToolTip,
  Truncate,
} from '@/components'
import {
  getComponentBuild,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponentOutputs,
  getInstallDeploy,
  getInstallDeployPlan,
} from '@/lib'
import type { TInstallDeployPlan, TInstall } from '@/types'
import { CANCEL_RUNNER_JOBS, DEPLOY_INTERMEDIATE_DATA } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const componentId = params?.['component-id'] as string
  const deployId = params?.['deploy-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [deploy, component] = await Promise.all([
    getInstallDeploy({
      installDeployId: deployId,
      installId,
      orgId,
    }),
    getComponent({ componentId, orgId }),
  ])

  return {
    title: `${component.name} | ${deploy.install_deploy_type}`,
  }
}

export default withPageAuthRequired(async function InstallComponentDeploy({
  params,
}) {
  const componentId = params?.['component-id'] as string
  const deployId = params?.['deploy-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [component, deploy, install] = await Promise.all([
    getComponent({
      componentId,
      orgId,
    }),
    getInstallDeploy({
      installDeployId: deployId,
      installId,
      orgId,
    }),
    getInstall({ installId, orgId }),
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
          href: `/${orgId}/installs/${install.id}/components/${deploy.component_id}`,
          text: component.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/components/${deploy.component_id}/deploys/${deploy.id}`,
          text: deploy.id,
        },
      ]}
      heading={`${component.name} ${deploy.install_deploy_type}`}
      headingUnderline={deploy.id}
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlank />
            <Time time={deploy.created_at} />
          </Text>
          <Text>
            <Timer />
            <Duration
              beginTime={deploy.created_at}
              endTime={deploy.updated_at}
            />
          </Text>
        </div>
      }
      statues={
        <div className="flex gap-6 items-start justify-start">
          <span className="flex flex-col gap-2">
            <DeployStatus
              descriptionAlignment="right"
              initDeploy={deploy}
              shouldPoll
            />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Type
            </Text>
            <Text>{deploy.install_deploy_type}</Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Build
            </Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={deploy.build_id}>
                <ClickToCopy>
                  <Truncate variant="small">{deploy.build_id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Component
            </Text>
            <Text variant="med-12">{component.name}</Text>
            <Text variant="mono-12">
              <ToolTip alignment="right" tipContent={deploy.component_id}>
                <ClickToCopy>
                  <Truncate variant="small">{deploy.component_id}</Truncate>
                </ClickToCopy>
              </ToolTip>
            </Text>
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
          {CANCEL_RUNNER_JOBS &&
          deploy?.status !== 'active' &&
          deploy?.status !== 'error' &&
          deploy?.status !== 'inactive' &&
          deploy?.runner_jobs?.length ? (
            <CancelRunnerJobButton
              jobType="deploy"
              runnerJobId={deploy?.runner_jobs?.at(-1)?.id}
              orgId={orgId}
            />
          ) : null}
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          <LogStreamProvider initLogStream={deploy?.log_stream}>
            <OperationLogsSection heading="Deploy logs" />
          </LogStreamProvider>
        </div>

        <div className="divide-y flex flex-col md:col-span-4">
          <Section
            className="flex-initial"
            actions={
              <Text>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}/builds/${deploy.build_id}`}
                >
                  Details
                  <CaretRight />
                </Link>
              </Text>
            }
            heading="Component build"
          >
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading component build..."
                    variant="stack"
                  />
                }
              >
                <LoadComponentBuild buildId={deploy.build_id} orgId={orgId} />
              </Suspense>
            </ErrorBoundary>
          </Section>

          <Section
            className="flex-initial"
            actions={
              <Text>
                <Link
                  href={`/${orgId}/apps/${component.app_id}/components/${component.id}`}
                >
                  Details
                  <CaretRight />
                </Link>
              </Text>
            }
            heading="Component config"
            childrenClassName="flex flex-col gap-4"
          >
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Loading
                    loadingText="Loading component config..."
                    variant="stack"
                  />
                }
              >
                <LoadComponentConfig
                  buildId={deploy.build_id}
                  componentId={deploy.component_id}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>            
          </Section>

          {DEPLOY_INTERMEDIATE_DATA ? (
            <ErrorBoundary fallbackRender={ErrorFallback}>
              <Suspense
                fallback={
                  <Section>
                    <Loading
                      loadingText="Loading intermediate data..."
                      variant="stack"
                    />
                  </Section>
                }
              >
                <LoadIntermediateData
                  deployId={deploy.id}
                  install={install}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          ) : null}
        </div>
      </div>
    </DashboardContent>
  )
})

// load log stream

// load component build
const LoadComponentBuild: FC<{ buildId: string; orgId: string }> = async ({
  buildId,
  orgId,
}) => {
  const build = await getComponentBuild({ buildId, orgId }).catch(console.error)

  return build ? (
    <div className="flex items-start justify-start gap-6">
      <span className="flex flex-col gap-2">
        <StatusBadge
          description={build.status_description}
          status={build.status}
          label="Status"
        />
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Build date
        </Text>
        <Time time={build.created_at} />
      </span>

      <span className="flex flex-col gap-2">
        <Text className="text-cool-grey-600 dark:text-cool-grey-500">
          Build duration
        </Text>
        <Duration beginTime={build.created_at} endTime={build.updated_at} />
      </span>
    </div>
  ) : (
    <Text>No component build found.</Text>
  )
}

// load component config
const LoadComponentConfig: FC<{
  componentId: string
  buildId: string
  orgId: string
}> = async ({ componentId, buildId, orgId }) => {
  const build = await getComponentBuild({ buildId, orgId })
  const componentConfig = await getComponentConfig({
    componentId,
    componentConfigId: build.component_config_connection_id,
    orgId,
  }).catch(console.error)
  return componentConfig ? (
    <ComponentConfiguration config={componentConfig} />
  ) : (
    <Text>No component config found.</Text>
  )
}

// load intermediate data
const LoadIntermediateData: FC<{
  deployId: string
  install: TInstall
  orgId: string
}> = async ({ deployId, install, orgId }) => {
  const deployPlan = await getInstallDeployPlan({
    deployId,
    installId: install.id,
    orgId,
  }).catch(console.error)

  return deployPlan &&
    (deployPlan as TInstallDeployPlan)?.actual?.waypoint_plan?.variables
      ?.intermediate_data?.nuon ? (
    <Section
      childrenClassName="flex flex-col gap-8"
      heading="Rendered intermediate data"
      className="flex-initial"
    >
      <InstallDeployIntermediateData
        install={install}
        data={
          (deployPlan as TInstallDeployPlan)?.actual?.waypoint_plan?.variables
            ?.intermediate_data
        }
      />
    </Section>
  ) : null
}

// load latest output
const LoadLatestOutputs: FC<{
  componentId: string
  installId: string
  orgId: string
}> = async ({ componentId, installId, orgId }) => {
  const outputs = await getInstallComponentOutputs({
    componentId,
    installId,
    orgId,
  }).catch(console.error)

  return outputs ? (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Text variant="med-12">Outputs</Text>
        <ClickToCopyButton textToCopy={JSON.stringify(outputs)} />
      </div>
      <CodeViewer
        initCodeSource={JSON.stringify(outputs, null, 2)}
        language="json"
      />
    </div>
  ) : null
}
