import { useQuery } from '@tanstack/react-query'
import type { TPlanInstallFacts } from '@/components/branches/PlanInstallChanges'
import type { IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { getWorkflowStep } from '@/lib'
import type { TInstallWorkflowStep } from '@/types'
import { PlanDiffPanel } from './PlanDiffPanel'

interface IPlanDiffPanelContainer extends IPanel {
  step: TInstallWorkflowStep
  installFacts?: TPlanInstallFacts
  workflowId?: string
}

export const PlanDiffPanelContainer = ({
  step: initStep,
  workflowId,
  ...props
}: IPlanDiffPanelContainer) => {
  const { org } = useOrg()
  const resolvedWorkflowId = workflowId || initStep.install_workflow_id

  const { data: step = initStep } = useQuery<TInstallWorkflowStep>({
    queryKey: ['workflow-step', org?.id, resolvedWorkflowId, initStep.id],
    queryFn: () =>
      getWorkflowStep({
        orgId: org!.id,
        workflowId: resolvedWorkflowId!,
        workflowStepId: initStep.id!,
      }),
    initialData: initStep,
    refetchInterval: (query) => (query.state.data?.finished ? false : 5000),
    enabled: !!org?.id && !!resolvedWorkflowId && !!initStep.id,
  })

  return (
    <PlanDiffPanel
      {...props}
      step={step}
      orgId={org?.id ?? ''}
      workflowId={resolvedWorkflowId}
    />
  )
}
