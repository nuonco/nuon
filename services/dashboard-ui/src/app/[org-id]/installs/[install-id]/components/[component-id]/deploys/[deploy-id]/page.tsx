import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  CalendarBlank,
  CaretLeft,
  CaretRight,
  Timer,
} from '@phosphor-icons/react/dist/ssr'
import {
  ApprovalStep,
  ClickToCopy,
  ComponentConfiguration,
  ConfigurationVariables,
  DashboardContent,
  DeployStatus,
  Duration,
  ErrorFallback,
  InstallComponentManagementDropdown,
  InstallWorkflowCancelModal,
  Link,
  Loading,
  LogStreamProvider,
  OperationLogsSection,
  RunnerJobPlanModal,
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
  getInstall,
  getInstallDeploy,
  getInstallWorkflow,
} from '@/lib'
import type { TBuild, TComponentConfig } from '@/types'
import { CANCEL_RUNNER_JOBS, sizeToMbOrGB, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
    ['deploy-id']: deployId,
  } = await params
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

export default async function InstallComponentDeploy({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
    ['deploy-id']: deployId,
  } = await params
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

  const installWorkflow = await getInstallWorkflow({
    installWorkflowId: deploy?.install_workflow_id,
    orgId,
  }).catch(console.error)
  const step = installWorkflow
    ? installWorkflow?.steps?.find((s) => s?.step_target_id === deploy?.id)
    : null

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install?.id}/components`,
          text: 'Components',
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
      headingMeta={
        deploy?.install_workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/workflows/${deploy?.install_workflow_id}?target=${deployId}`}
          >
            <CaretLeft />
            View workflow
          </Link>
        ) : null
      }
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
        <div className="flex gap-6 items-start justify-end flex-wrap">
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

          <div className="flex flex-wrap gap-4 items-center">
            <ErrorBoundary
              fallback={<Text>Can&apso;t fetching sync plan</Text>}
            >
              <Suspense
                fallback={
                  <Loading variant="stack" loadingText="Loading sync plan..." />
                }
              >
                <RunnerJobPlanModal
                  buttonText="Sync plan"
                  headingText="Component sync plan"
                  runnerJobId={deploy?.runner_jobs?.at(-1)?.id}
                />
              </Suspense>
            </ErrorBoundary>

            {deploy?.install_deploy_type !== 'sync-image' ? (
              <ErrorBoundary
                fallback={<Text>Can&apso;t fetching deploy plan</Text>}
              >
                <Suspense
                  fallback={
                    <Loading
                      variant="stack"
                      loadingText="Loading deploy plan..."
                    />
                  }
                >
                  <RunnerJobPlanModal
                    buttonText="Deploy plan"
                    headingText="Component deploy plan"
                    runnerJobId={deploy?.runner_jobs?.at(0)?.id}
                  />
                </Suspense>
              </ErrorBoundary>
            ) : null}

            {component ? (
              <InstallComponentManagementDropdown component={component} />
            ) : null}
            {CANCEL_RUNNER_JOBS &&
            deploy?.status !== 'active' &&
            deploy?.status !== 'error' &&
            deploy?.status !== 'inactive' &&
            deploy?.runner_jobs?.length &&
            installWorkflow &&
            !installWorkflow?.finished ? (
              <InstallWorkflowCancelModal installWorkflow={installWorkflow} />
            ) : null}
          </div>
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {installWorkflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped'? (
            <Section
              className="border-b"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={installWorkflow?.id}
              />
            </Section>
          ) : null}

          <LogStreamProvider initLogStream={deploy?.log_stream}>
            <OperationLogsSection heading="Deploy logs" />
          </LogStreamProvider>

          {installWorkflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped'? (
            <Section
              className="border-t"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={installWorkflow?.id}
              />
            </Section>
          ) : null}
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

          {deploy?.oci_artifact ? (
            <Section>
              <ConfigurationVariables
                heading="OCI artifacts"
                headingVariant="semi-14"
                isNotTruncated
                variables={{
                  tag: deploy?.oci_artifact?.tag,
                  repository: deploy?.oci_artifact?.repository,
                  digest: deploy?.oci_artifact?.digest,
                  size: sizeToMbOrGB(deploy?.oci_artifact?.size),
                  artifact_type: deploy?.oci_artifact?.artifact_type,
                  urls: deploy?.oci_artifact?.urls as unknown as string,
                }}
              />
            </Section>
          ) : null}
        </div>
      </div>
    </DashboardContent>
  )
}

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
  const { data: build, error: buildError } = await nueQueryData<TBuild>({
    orgId,
    path: `components/builds/${buildId}`,
  })

  const { data: componentConfig, error } = await nueQueryData<TComponentConfig>(
    {
      orgId,
      path: `components/${componentId}/configs/${build?.component_config_connection_id}`,
    }
  )

  return buildError || error ? (
    <Text>{buildError?.error || error?.error}</Text>
  ) : componentConfig ? (
    <ComponentConfiguration config={componentConfig} hideHelmValuesFile />
  ) : (
    <Text>No component config found.</Text>
  )
}
