import { useQuery } from '@tanstack/react-query'
import { getDeploy, getInstallActionRun, getInstallComponentOutputs } from '@/lib'
import type { TDeploy, TWorkflowStep } from '@/types'
import { RunbookStepCard } from './RunbookStepCard'

export interface IRunbookStepCardContainer {
  step: TWorkflowStep
  installId: string
  orgId: string
  workflowUrl: string
}

export const RunbookStepCardContainer = ({
  step,
  installId,
  orgId,
  workflowUrl,
}: IRunbookStepCardContainer) => {
  const isDeploy = step.step_target_type === 'install_deploys'
  const isActionRun = step.step_target_type === 'install_action_workflow_runs'
  const targetId = step.step_target_id

  const { data, isLoading } = useQuery({
    queryKey: ['runbook-step-data', step.id, targetId],
    queryFn: () => {
      if (isDeploy) {
        return getDeploy({ installId, deployId: targetId!, orgId })
      }
      if (isActionRun) {
        return getInstallActionRun({ installId, runId: targetId!, orgId })
      }
      return Promise.resolve(null)
    },
    enabled: !!targetId && (isDeploy || isActionRun),
  })

  const componentId = isDeploy ? (data as TDeploy)?.component_id : undefined

  const { data: deployOutputs } = useQuery({
    queryKey: ['runbook-step-deploy-outputs', step.id, componentId],
    queryFn: () =>
      getInstallComponentOutputs({
        orgId,
        installId,
        componentId: componentId!,
      }),
    enabled: isDeploy && !!componentId,
    retry: false,
  })

  return (
    <RunbookStepCard
      step={step}
      workflowUrl={workflowUrl}
      targetData={data}
      deployOutputs={deployOutputs}
      isLoading={isLoading}
    />
  )
}
