import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { InstallActionManualRunButton } from '@/components/actions/InstallActionManualRun'
import { Icon } from '@/components/common/Icon'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { CancelRunnerJobButton } from '@/components/runners/CancelRunnerJob'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { useOrg } from '@/hooks/use-org'
import { useAuth } from '@/hooks/use-auth'
import type { TInstallAction, TWorkflow } from '@/types'
import { getWorkflowStep } from '@/utils/workflow-utils'
import { InstallActionRunHeader } from './InstallActionRunHeader'

interface IInstallActionRunHeaderContainer {
  action?: TInstallAction
  actionId: string
  actionName: string
  workflow: TWorkflow
}

export const InstallActionRunHeaderContainer = ({
  action,
  actionId,
  actionName,
  workflow,
}: IInstallActionRunHeaderContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { installActionRun } = useInstallActionRun()
  const { isAdmin } = useAuth()
  const step = getWorkflowStep({
    workflow,
    stepTargetId: installActionRun?.id,
  })
  const basePath = `/${org?.id}/installs/${install?.id}`

  const hasWorkflow = !!installActionRun?.install_workflow_id
  const runnerJobStatus = installActionRun?.runner_job?.status
  const isRunnerJobActive =
    runnerJobStatus === 'queued' ||
    runnerJobStatus === 'in-progress' ||
    runnerJobStatus === 'available'
  const cancelButton = hasWorkflow ? (
    <CancelWorkflowButton workflow={workflow} />
  ) : installActionRun?.runner_job && isRunnerJobActive ? (
    <CancelRunnerJobButton runnerJob={installActionRun.runner_job} jobType="actions" />
  ) : null

  const actionWorkflow = action?.action_workflow
  const hasManualTrigger =
    actionWorkflow?.configs?.[0]?.triggers?.find((t) => t.type === 'manual') ||
    installActionRun?.triggered_by_type === 'manual'

  return (
    <InstallActionRunHeader
      actionId={actionId}
      actionName={actionName}
      workflow={workflow}
      installActionRun={installActionRun}
      basePath={basePath}
      isAdmin={isAdmin}
      step={step}
      cancelWorkflowButton={cancelButton}
      runActionButton={
        hasManualTrigger && actionWorkflow ? (
          <InstallActionManualRunButton
            action={actionWorkflow}
            actionConfigId={
              actionWorkflow.configs?.[0]?.id ??
              installActionRun?.action_workflow_config_id ??
              ''
            }
            variant="primary"
          >
            Re-run action
            <Icon variant="PlayIcon" />
          </InstallActionManualRunButton>
        ) : null
      }
      runnerJobPlanButton={
        installActionRun?.runner_job?.id ? (
          <RunnerJobPlanButton
            runnerJobId={installActionRun?.runner_job?.id}
          />
        ) : null
      }
    />
  )
}
