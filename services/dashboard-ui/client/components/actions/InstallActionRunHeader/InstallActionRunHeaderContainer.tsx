import { RunnerJobPlanButton } from '@/components/runners/RunnerJobPlan'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import { CancelRunnerJobButton } from '@/components/runners/CancelRunnerJob'
import { useInstall } from '@/hooks/use-install'
import { useInstallActionRun } from '@/hooks/use-install-action-run'
import { useOrg } from '@/hooks/use-org'
import { useAuth } from '@/hooks/use-auth'
import type { TWorkflow } from '@/types'
import { getWorkflowStep } from '@/utils/workflow-utils'
import { InstallActionRunHeader } from './InstallActionRunHeader'

interface IInstallActionRunHeaderContainer {
  actionId: string
  actionName: string
  workflow: TWorkflow
}

export const InstallActionRunHeaderContainer = ({
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
  const cancelButton = hasWorkflow ? (
    <CancelWorkflowButton workflow={workflow} />
  ) : installActionRun?.runner_job ? (
    <CancelRunnerJobButton runnerJob={installActionRun.runner_job} jobType="actions" />
  ) : null

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
