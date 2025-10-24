import type { Metadata } from 'next'
import { Suspense } from 'react'
import {
  CalendarBlankIcon,
  CaretLeftIcon,
  TimerIcon,
} from '@phosphor-icons/react/dist/ssr'

import { ActionStepGraph } from '@/components/actions/ActionStepsGraph'
import { CodeBlock } from '@/components/common/CodeBlock'
import { Text } from '@/components/common/Text'

import {
  getInstallActionById,
  getInstallActionRunById,
  getInstallById,
  getWorkflowById,
  getOrgById,
} from '@/lib'
import { InstallActionRunProvider } from '@/providers/install-action-run-provider'
import { hydrateActionRunSteps } from '@/utils/action-utils'
import { CANCEL_RUNNER_JOBS } from '@/utils'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  ActionTriggerType,
  ActionLogsSection,
  ActionWorkflowStatus,
  ClickToCopy,
  DashboardContent,
  Duration,
  Link as OldLink,
  Loading,
  LogStreamProvider,
  RunnerJobPlanModal,
  Text as OldText,
  Time,
  ToolTip,
} from '@/components'
import { InstallActionCancelButton } from '@/components/old/InstallActionRunCancelButton'
import { InstallActionRunDetails } from '@/components/old/InstallActionRunDetails'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [{ data: installActionRun }, { data: installAction }] =
    await Promise.all([
      getInstallActionRunById({
        installId,
        orgId,
        runId,
      }),
      getInstallActionById({
        actionId,
        installId,
        orgId,
      }),
    ])

  return {
    title: `${installAction?.action_workflow?.name} | ${installActionRun.trigger_type} run`,
  }
}

export default async function InstallActionRunPage({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['action-id']: actionId,
    ['run-id']: runId,
  } = await params
  const [
    { data: install },
    { data: installActionRun },
    { data: installAction },
    { data: org },
  ] = await Promise.all([
    getInstallById({ installId, orgId }),
    getInstallActionRunById({
      installId,
      orgId,
      runId,
    }),
    getInstallActionById({
      actionId,
      installId,
      orgId,
    }),
    getOrgById({ orgId }),
  ])

  const { data: workflow } = await getWorkflowById({
    orgId,
    workflowId: installActionRun?.install_workflow_id,
  })
  const step = workflow
    ? workflow?.steps
        ?.filter((s) => s?.step_target_id === installActionRun?.id)
        ?.at(-1)
    : null

  const actionConfig = installAction?.action_workflow?.configs?.find(
    (cfg) => cfg?.id === installActionRun?.action_workflow_config_id
  )

  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun}
      shouldPoll
    >
      {org?.features?.['stratus-layout'] ? (
        <div className="flex flex-col gap-6">
          <ActionStepGraph
            steps={hydrateActionRunSteps({
              steps: installActionRun?.steps,
              stepConfigs: actionConfig?.steps,
            })}
          />

          <Text variant="body" weight="strong">
            Outputs
          </Text>
          <CodeBlock language="json">
            {JSON.stringify(installActionRun?.runner_job?.outputs, null, 2)}
          </CodeBlock>
        </div>
      ) : (
        <DashboardContent
          breadcrumb={[
            { href: `/${orgId}/installs`, text: 'Installs' },
            {
              href: `/${orgId}/installs/${install.id}`,
              text: install.name,
            },
            {
              href: `/${orgId}/installs/${install.id}/actions`,
              text: 'Actions',
            },
            {
              href: `/${orgId}/installs/${install.id}/actions/${actionId}`,
              text: `${installAction?.action_workflow?.name}`,
            },
            {
              href: `/${orgId}/installs/${install.id}/actions/${actionId}/${installActionRun.id}`,
              text: installActionRun.id,
            },
          ]}
          heading={`${installAction?.action_workflow?.name} execution`}
          headingUnderline={actionId}
          headingMeta={
            installActionRun?.install_workflow_id ? (
              <OldLink
                href={`/${orgId}/installs/${installId}/workflows/${installActionRun?.install_workflow_id}?target=${step?.id}`}
              >
                <CaretLeftIcon />
                View workflow
              </OldLink>
            ) : null
          }
          meta={
            <div className="flex gap-8 items-center justify-start pb-6">
              <OldText>
                <CalendarBlankIcon size={14} />
                <Time time={installActionRun.created_at} />
              </OldText>
              <OldText>
                <TimerIcon size={14} />
                <Duration nanoseconds={installActionRun.execution_time} />
              </OldText>
            </div>
          }
          statues={
            <div className="flex gap-6 items-start justify-start">
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Status
                </OldText>
                <ActionWorkflowStatus descriptionAlignment="right" />
              </span>
              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Trigger type
                </OldText>
                <ActionTriggerType
                  triggerType={installActionRun?.triggered_by_type}
                  componentName={installActionRun?.run_env_vars?.COMPONENT_NAME}
                  componentPath={`/${orgId}/installs/${installId}/components/${installActionRun?.run_env_vars?.COMPONENT_ID}`}
                />
              </span>

              <span className="flex flex-col gap-2">
                <OldText className="text-cool-grey-600 dark:text-cool-grey-500">
                  Install
                </OldText>
                <OldText variant="med-12">{install.name}</OldText>
                <OldText variant="mono-12">
                  <ToolTip alignment="right" tipContent={install.id}>
                    <ClickToCopy>{install.id}</ClickToCopy>
                  </ToolTip>
                </OldText>
              </span>
              {installActionRun?.runner_job?.id ? (
                <ErrorBoundary
                  fallback={<OldText>Can&apso;t fetching job plan</OldText>}
                >
                  <Suspense
                    fallback={
                      <Loading
                        variant="stack"
                        loadingText="Loading job plan..."
                      />
                    }
                  >
                    <RunnerJobPlanModal
                      runnerJobId={installActionRun?.runner_job?.id}
                    />
                  </Suspense>
                </ErrorBoundary>
              ) : null}
              {CANCEL_RUNNER_JOBS ? (
                <InstallActionCancelButton workflow={workflow} />
              ) : null}
            </div>
          }
        >
          <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x">
            <div className="md:col-span-8">
              <LogStreamProvider initLogStream={installActionRun?.log_stream}>
                <ActionLogsSection />
              </LogStreamProvider>
            </div>
            <InstallActionRunDetails />
          </div>
        </DashboardContent>
      )}
    </InstallActionRunProvider>
  )
}
