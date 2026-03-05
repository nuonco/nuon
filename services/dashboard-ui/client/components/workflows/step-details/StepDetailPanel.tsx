import { useSearchParams } from 'react-router'
import React, { useEffect, type ReactElement, type ReactNode } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { getWorkflowStep } from '@/lib'
import type { TWorkflowStep } from '@/types'
import { ActionRunStepDetails } from './action-run-details/ActionRunStepDetails'
import { DeployStepDetails } from './deploy-details/DeployStepDetails'
import { SandboxRunStepDetails } from './sandbox-run-details/SandboxRunStepDetails'
import { StackStepDetails } from './stack-details/StackStepDetails'
import { StepBanner } from './StepBanner'
import { StepTitle } from './StepTitle'
import { StepMetadata } from './StepMetadata'
import { RunnerStepDetails } from './RunnerStepDetails'

type TPanelSize = IPanel['size']

function getStepPanelSize(step: TWorkflowStep): TPanelSize {
  if (
    step?.step_target_type === 'install_deploys' ||
    step?.step_target_type === 'install_sandbox_runs' ||
    step?.step_target_type === 'install_action_workflow_runs'
  ) {
    return '3/4'
  }
  return 'half'
}

function getStepPanelDetails(step: TWorkflowStep): ReactNode {
  if (step.step_target_type === 'install_action_workflow_runs') return <ActionRunStepDetails />
  if (step.step_target_type === 'install_deploys') return <DeployStepDetails />
  if (step.step_target_type === 'install_sandbox_runs') return <SandboxRunStepDetails />
  if (step.step_target_type === 'install_stack_versions') return <StackStepDetails />
  if (step.step_target_type === 'runners') return <RunnerStepDetails />
}

export interface IStepDetailPanel extends IPanel {
  children: ReactNode
  initStep: TWorkflowStep
  planOnly?: boolean
  pollInterval?: number
  shouldPoll?: boolean
}

export const StepDetailPanel = ({
  children,
  initStep,
  planOnly = false,
  pollInterval = 10000,
  shouldPoll = false,
  ...props
}: IStepDetailPanel) => {
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
    <Panel
      className="@container"
      heading={<StepTitle step={step} />}
      size="half"
      {...props}
    >
      <StepBanner step={step} planOnly={planOnly} />
      {React.Children.map(children, (c) =>
        React.isValidElement(c)
          ? React.cloneElement(
              c as ReactElement<{ step: TWorkflowStep; panelId: string }>,
              { step, panelId: props.panelId }
            )
          : null
      )}

      <Divider dividerWord="Metadata" />

      <StepMetadata step={step} />
    </Panel>
  )
}

export const StepDetailPanelButton = ({
  step,
  planOnly = false,
}: {
  approvalPrompt?: boolean
  step: TWorkflowStep
  planOnly?: boolean
}) => {
  const { addPanel } = useSurfaces()
  const [searchParams] = useSearchParams()

  const panel = (
    <StepDetailPanel
      panelKey={step.id}
      initStep={step}
      size={getStepPanelSize(step)}
      shouldPoll
      planOnly={planOnly}
    >
      {getStepPanelDetails(step)}
    </StepDetailPanel>
  )

  const handleAddPanel = () => addPanel(panel, step.id)

  useEffect(() => {
    if (step.id && step.id === searchParams?.get('panel')) {
      handleAddPanel()
    }
  }, [])

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
