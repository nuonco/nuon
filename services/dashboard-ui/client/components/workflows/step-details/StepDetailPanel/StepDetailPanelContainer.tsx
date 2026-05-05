import { useSearchParams } from 'react-router'
import { useEffect, useRef, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Icon } from '@/components/common/Icon'
import type { IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { useWorkflow } from '@/hooks/use-workflow'
import { getWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { ActionRunStepDetails } from '../action-run-details/ActionRunStepDetails'
import { DeployStepDetails } from '../deploy-details/DeployStepDetails'
import { SandboxRunStepDetails } from '../sandbox-run-details/SandboxRunStepDetails'
import { StackStepDetails } from '../stack-details/StackStepDetails'
import { RunnerStepDetails } from '../RunnerStepDetails'
import { SyncSecretsStepDetails } from '../SyncSecretsStepDetails'
import { StepDetailPanel } from './StepDetailPanel'

type TPanelSize = IPanel['size']

export function getStepPanelSize(step: TWorkflowStep): TPanelSize {
  if (
    step?.step_target_type === 'install_deploys' ||
    step?.step_target_type === 'install_sandbox_runs' ||
    step?.step_target_type === 'install_action_workflow_runs' ||
    step?.step_target_type === 'install_workflow_steps'
  ) {
    return '3/4'
  }
  return 'half'
}

export function getStepPanelDetails(step: TWorkflowStep): ReactNode {
  if (step.step_target_type === 'install_action_workflow_runs') return <ActionRunStepDetails />
  if (step.step_target_type === 'install_deploys') return <DeployStepDetails />
  if (step.step_target_type === 'install_sandbox_runs') return <SandboxRunStepDetails />
  if (step.step_target_type === 'install_stack_versions') return <StackStepDetails />
  if (step.step_target_type === 'runners') return <RunnerStepDetails />
  if (step.step_target_type === 'install_workflow_steps') return <SyncSecretsStepDetails />
}

export interface IStepDetailPanelContainer extends IPanel {
  children: ReactNode
  initStep: TWorkflowStep
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
}

export const StepDetailPanelContainer = ({
  children,
  initStep,
  planOnly = false,
  pollInterval = 10000,
  shouldPoll = false,
  ...props
}: IStepDetailPanelContainer) => {
  const { org } = useOrg()

  const { data: step = initStep } = useQuery<TWorkflowStep>({
    queryKey: ['workflow-step', org?.id, initStep.install_workflow_id, initStep.id],
    queryFn: () =>
      getWorkflowStep({
        orgId: org.id,
        workflowId: initStep.install_workflow_id,
        workflowStepId: initStep.id,
      }),
    refetchInterval: shouldPoll ? pollInterval : false,
    initialData: initStep,
    enabled: !!org?.id,
  })

  return (
    <StepDetailPanel
      step={step}
      planOnly={planOnly}
      {...props}
    >
      {children}
    </StepDetailPanel>
  )
}

export const StepDetailPanelButton = ({
  step,
  approvalPrompt,
  planOnly = false,
}: {
  approvalPrompt?: boolean
  step: TWorkflowStep
  planOnly?: boolean
}) => {
  const { addPanel, removePanel, panels } = useSurfaces()
  const { workflow } = useWorkflow()
  const [searchParams] = useSearchParams()
  const autoOpened = useRef(false)
  const panelIdRef = useRef<string | null>(null)

  const panel = (
    <StepDetailPanelContainer
      panelKey={step.id}
      initStep={step}
      size={getStepPanelSize(step)}
      shouldPoll
      planOnly={planOnly}
    >
      {getStepPanelDetails(step)}
    </StepDetailPanelContainer>
  )

  const handleAddPanel = () => {
    panelIdRef.current = addPanel(panel, step.id)
  }

  const isPendingApproval =
    approvalPrompt &&
    step.execution_type === 'approval' &&
    !!step.approval &&
    !step.approval.response &&
    step.status?.status !== 'discarded'

  const isPendingAwaitStack =
    step.step_target_type === 'install_stack_versions' &&
    step.name !== 'generate install stack' &&
    !!step.started_at &&
    !step.finished

  const workflowStatus = workflow?.status?.status
  const workflowBlocked = workflowStatus === 'cancelled' || workflowStatus === 'error'
  const suppressAutoOpen = planOnly || workflow?.type === 'drift_run' || workflow?.type === 'drift_run_reprovision_sandbox'

  useEffect(() => {
    if (!suppressAutoOpen && step.id && step.id === searchParams?.get('panel')) {
      handleAddPanel()
    }
  }, [])

  useEffect(() => {
    if (autoOpened.current || !workflow) return
    if (!workflowBlocked && !suppressAutoOpen && (isPendingApproval || isPendingAwaitStack) && panels.length === 0) {
      autoOpened.current = true
      handleAddPanel()
    }
  }, [workflow?.id, workflowBlocked, suppressAutoOpen, isPendingApproval, isPendingAwaitStack])

  useEffect(() => {
    if (autoOpened.current && panelIdRef.current && step.finished) {
      removePanel(panelIdRef.current, step.id)
      panelIdRef.current = null
    }
  }, [step.finished])

  return (
    <Button
      className="!text-primary-600 dark:!text-primary-500"
      variant="ghost"
      size="sm"
      onClick={handleAddPanel}
    >
      View details <Icon variant="CaretRight" />
    </Button>
  )
}
