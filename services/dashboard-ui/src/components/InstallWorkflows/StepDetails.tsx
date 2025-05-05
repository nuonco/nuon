import React, { type FC } from 'react'
import { Loading } from '@/components/Loading'
import { Text, Code } from '@/components/Typography'
import type { TInstallWorkflowStep, TInstall } from '@/types'
import { ActionStepDetails } from './ActionStepDetails'
import { DeployStepDetails } from './DeployStepDetails'
import { SandboxStepDetails } from './SandboxStepDetails'
import { StackStep } from './StackStepDetails'
import { RunnerStepDetails } from './RunnerStepDetails'

export function getStepType(
  step: TInstallWorkflowStep,
  install: TInstall
): React.ReactNode {
  let stepDetails = <>Unknown step</>
  switch (step.step_target_type) {
    case 'install_sandbox_runs':
      stepDetails = (
        <SandboxStepDetails
          step={step}
          shouldPoll={step?.status?.status === 'in-progress'}
        />
      )
      break

    case 'install_stack_versions':
      stepDetails = (
        <StackStep
          step={step}
          appId={install?.app_id}
          shouldPoll={step?.status?.status === 'in-progress'}
        />
      )
      break

    case 'install_action_workflow_runs':
      stepDetails = (
        <ActionStepDetails
          step={step}
          shouldPoll={step?.status?.status === 'in-progress'}
        />
      )
      break

    case 'runners':
      stepDetails = (
        <RunnerStepDetails
          step={step}
          shouldPoll={step?.status?.status === 'in-progress'}
        />
      )
      break
    case 'install_deploys':
      stepDetails = (
        <DeployStepDetails
          step={step}
          shouldPoll={step?.status?.status === 'in-progress'}
        />
      )
      break
    default:
      stepDetails = <Loading loadingText="Waiting on step..." variant="page" />
  }

  return (
    <>
      {stepDetails}
      <Text>Step JSON</Text>
      <Code variant="preformated">{JSON.stringify(step, null, 2)}</Code>
    </>
  )
}

export const StepDetails: FC<{
  children: React.ReactNode
  activeStepIndex: number
}> = ({ children, activeStepIndex = 0 }) => {
  const steps = React.Children.toArray(children)

  return <div>{steps[activeStepIndex]}</div>
}
