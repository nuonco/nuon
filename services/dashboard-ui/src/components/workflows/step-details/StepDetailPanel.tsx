'use client'

import { useSearchParams } from 'next/navigation'
import React, { useEffect, type ReactElement, type ReactNode } from 'react'
import { Button } from '@/components/common/Button'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { Panel, type IPanel } from '@/components/surfaces/Panel'
import { useOrg } from '@/hooks/use-org'
import { useSurfaces } from '@/hooks/use-surfaces'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TWorkflowStep } from '@/types'
import { DeployStepDetails } from './deploy-details/DeployStepDetails'
import { SandboxRunStepDetails } from './sandbox-run-details/SandboxRunStepDetails'
import { StackStepDetails } from './stack-details/StackStepDetails'
import { StepBanner } from './StepBanner'
import { StepTitle } from './StepTitle'
import { StepMetadata } from './StepMetadata'
import { ActionRunStepDetails } from './ActionRunStepDetails'
import { RunnerStepDetails } from './RunnerStepDetails'

function getStepPanelDetails(step: TWorkflowStep): ReactNode {
  if (step.step_target_type === 'install_action_workflow_runs') {
    return <ActionRunStepDetails />
  }

  if (step.step_target_type === 'install_deploys') {
    return <DeployStepDetails />
  }

  if (step.step_target_type === 'install_sandbox_runs') {
    return <SandboxRunStepDetails />
  }

  if (step.step_target_type === 'install_stack_versions') {
    return <StackStepDetails />
  }

  if (step.step_target_type === 'runners') {
    return <RunnerStepDetails />
  }
}

export interface IStepDetailPanel extends IPanel, IPollingProps {
  children: ReactNode
  initStep: TWorkflowStep
}

export const StepDetailPanel = ({
  children,
  initStep,
  pollInterval = 10000,
  shouldPoll = false,
  ...props
}: IStepDetailPanel) => {
  const { org } = useOrg()
  const { data: step } = usePolling<TWorkflowStep>({
    initData: initStep,
    path: `/api/orgs/${org.id}/workflows/${initStep.install_workflow_id}/steps/${initStep.id}`,
    pollInterval,
    shouldPoll,
  })

  return (
    <Panel
      className="@container"
      heading={<StepTitle step={step} />}
      size="half"
      {...props}
    >
      <StepBanner step={step} />
      {React.Children.map(children, (c) =>
        React.isValidElement(c)
          ? React.cloneElement(
              c as ReactElement<{ step: TWorkflowStep; panelId: string }>,
              {
                step,
                panelId: props.panelId,
              }
            )
          : null
      )}

      <Divider dividerWord="Metadata" />

      <StepMetadata step={step} />
    </Panel>
  )
}

export const StepDetailPanelButton = ({ step }: { step: TWorkflowStep }) => {
  const { addPanel } = useSurfaces()
  const searchParams = useSearchParams()
  const panel = (
    <StepDetailPanel
      panelKey={step.id}
      initStep={step}
      size={
        step?.step_target_type === 'install_deploys' &&
        step?.execution_type !== 'approval'
          ? 'full'
          : 'half'
      }
      shouldPoll
    >
      {getStepPanelDetails(step)}
    </StepDetailPanel>
  )

  const handleAddPanel = () => {
    addPanel(panel, step.id)
  }

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
