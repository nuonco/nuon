import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  CalendarBlankIcon,
  CaretLeftIcon,
  CaretRightIcon,
  TimerIcon,
} from '@phosphor-icons/react/dist/ssr'
import {
  ApprovalStep,
  ClickToCopy,
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
  Section,
  Text,
  Time,
  ToolTip,
  Truncate,
} from '@/components'
import { getInstallById, getDeployById, getWorkflowById } from '@/lib'
import { CANCEL_RUNNER_JOBS, sizeToMbOrGB } from '@/utils'
import { Build } from './build'
import { ComponentConfig } from './config'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['deploy-id']: deployId,
  } = await params
  const { data: deploy } = await getDeployById({
    deployId,
    installId,
    orgId,
  })

  return {
    title: `${deploy?.install_deploy_type} | ${deploy?.component_name} | Nuon`,
  }
}

export default async function InstallComponentDeploy({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['component-id']: componentId,
    ['deploy-id']: deployId,
  } = await params
  const [{ data: deploy, error, status }, { data: install }] =
    await Promise.all([
      getDeployById({
        deployId,
        installId,
        orgId,
      }),
      getInstallById({ installId, orgId }),
    ])

  if (error) {
    console.error(
      'Error rendering install deploy page: ',
      `API status: ${status}`,
      error
    )
    if (status === 404) {
      notFound()
    } else {
      // TODO(nnnat): show error message
      notFound()
    }
  }

  const { data: workflow } = await getWorkflowById({
    workflowId: deploy?.install_workflow_id,
    orgId,
  })
  const step = workflow
    ? workflow?.steps
        ?.filter(
          (s) =>
            s?.step_target_id === deploy?.id && s?.execution_type === 'approval'
        )
        ?.at(-1)
    : null

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${installId}`,
          text: install?.name,
        },
        {
          href: `/${orgId}/installs/${install?.id}/components`,
          text: 'Components',
        },
        {
          href: `/${orgId}/installs/${install?.id}/components/${componentId}`,
          text: deploy?.component_name,
        },
        {
          href: `/${orgId}/installs/${install?.id}/components/${componentId}/deploys/${deploy?.id}`,
          text: deploy?.id,
        },
      ]}
      heading={`${deploy?.component_name} ${deploy.install_deploy_type}`}
      headingUnderline={deploy.id}
      headingMeta={
        deploy?.install_workflow_id ? (
          <Link
            href={`/${orgId}/installs/${installId}/workflows/${deploy?.install_workflow_id}?target=${step?.id}`}
          >
            <CaretLeftIcon />
            View workflow
          </Link>
        ) : null
      }
      meta={
        <div className="flex gap-8 items-center justify-start pb-6">
          <Text>
            <CalendarBlankIcon />
            <Time time={deploy?.created_at} />
          </Text>
          <Text>
            <TimerIcon />
            <Duration
              beginTime={deploy?.created_at}
              endTime={deploy?.updated_at}
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
            <Text variant="med-12">{deploy?.component_name}</Text>
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

            {deploy?.component_name ? (
              <InstallComponentManagementDropdown
                componentId={deploy?.component_id}
                componentName={deploy?.component_name}
              />
            ) : null}
            {CANCEL_RUNNER_JOBS &&
            deploy?.status !== 'active' &&
            deploy?.status !== 'error' &&
            deploy?.status !== 'inactive' &&
            deploy?.runner_jobs?.length &&
            workflow &&
            !workflow?.finished ? (
              <InstallWorkflowCancelModal installWorkflow={workflow} />
            ) : null}
          </div>
        </div>
      }
    >
      <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
        <div className="md:col-span-8">
          {workflow &&
          step &&
          step?.approval &&
          !step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-b"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
              />
            </Section>
          ) : null}

          <LogStreamProvider initLogStream={deploy?.log_stream}>
            <OperationLogsSection heading="Deploy logs" />
          </LogStreamProvider>

          {workflow &&
          step &&
          step?.approval &&
          step?.approval?.response &&
          step?.status?.status !== 'auto-skipped' ? (
            <Section
              className="border-t"
              childrenClassName="flex flex-col gap-6"
              heading="Approve change"
            >
              <ApprovalStep
                step={step}
                approval={step.approval}
                workflowId={workflow?.id}
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
                  href={`/${orgId}/apps/${install?.app_id}/components/${componentId}/builds/${deploy.build_id}`}
                >
                  Details
                  <CaretRightIcon />
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
                <Build
                  buildId={deploy.build_id}
                  componentId={componentId}
                  orgId={orgId}
                />
              </Suspense>
            </ErrorBoundary>
          </Section>

          <Section
            className="flex-initial"
            actions={
              <Text>
                <Link
                  href={`/${orgId}/apps/${install?.app_id}/components/${componentId}`}
                >
                  Details
                  <CaretRightIcon />
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
                <ComponentConfig
                  appConfigId={install?.app_config_id}
                  appId={install?.app_id}
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
