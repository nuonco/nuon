import type { Metadata } from 'next'
import { Suspense } from 'react'
import {
  CalendarBlankIcon,
  CaretLeftIcon,
  TimerIcon,
} from '@phosphor-icons/react/dist/ssr'
import { ActionStepGraph } from '@/components/actions/ActionStepsGraph'
import { InstallActionRunOutputs } from '@/components/actions/InstallActionRunOutputs'
import { Text } from '@/components/common/Text'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import {
  getInstallAction,
  getInstallActionRun,
  getInstall,
  getWorkflow,
  getOrg,
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
      getInstallActionRun({
        installId,
        orgId,
        runId,
      }),
      getInstallAction({
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
    getInstall({ installId, orgId }),
    getInstallActionRun({
      installId,
      orgId,
      runId,
    }),
    getInstallAction({
      actionId,
      installId,
      orgId,
    }),
    getOrg({ orgId }),
  ])

  const { data: workflow } = await getWorkflow({
    orgId,
    workflowId: installActionRun?.install_workflow_id,
  })
  const step = workflow
    ? workflow?.steps
        ?.filter((s) => s?.step_target_id === installActionRun?.id)
        ?.at(-1)
    : null

  return (
    <InstallActionRunProvider
      initInstallActionRun={installActionRun}
      shouldPoll
    >
      <div className="flex flex-col gap-6">
        <Breadcrumbs
          breadcrumbs={[
            {
              path: `/${orgId}`,
              text: org?.name,
            },
            {
              path: `/${orgId}/installs`,
              text: 'Installs',
            },
            {
              path: `/${orgId}/installs/${installId}`,
              text: install?.name,
            },
            {
              path: `/${orgId}/installs/${installId}/actions`,
              text: 'Actions',
            },
            {
              path: `/${orgId}/installs/${installId}/actions/${actionId}`,
              text: installAction?.action_workflow?.name || 'Action',
            },
            {
              path: `/${orgId}/installs/${installId}/actions/${actionId}/${runId}`,
              text: `${installActionRun?.trigger_type} run`,
            },
          ]}
        />
        <ActionStepGraph
          steps={hydrateActionRunSteps({
            steps: installActionRun?.steps,
            stepConfigs: installActionRun?.config?.steps,
          })}
        />

        <Text variant="body" weight="strong">
          Outputs
        </Text>
        <InstallActionRunOutputs />
      </div>
    </InstallActionRunProvider>
  )
}
