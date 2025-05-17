import React, { type FC } from 'react'
import { Expand } from '@/components/Expand'
import { Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { Text, Code } from '@/components/Typography'
import type { TInstallWorkflowStep, TInstall } from '@/types'
import { sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'
import { ActionStepDetails } from './ActionStepDetails'
import { DeployStepDetails } from './DeployStepDetails'
import { SandboxStepDetails } from './SandboxStepDetails'
import { StackStep } from './StackStepDetails'
import { RunnerStepDetails } from './RunnerStepDetails'

export function getStepType(
  step: TInstallWorkflowStep,
  install: TInstall
): React.ReactNode {
  let stepDetails = <Loading loadingText="Waiting on step..." variant="page" />

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
      stepDetails = (
        <div className="flex flex-col gap-2">
          <Text variant="reg-14">
            Step doesn&apos;t have any details to show.
          </Text>
        </div>
      )
  }

  if (step?.execution_type === 'skipped') {
    stepDetails = (
      <div className="flex flex-col gap-2">
        <Text variant="reg-14">Step has been skipped</Text>
      </div>
    )
  }

  if (step?.execution_type === 'system' && step?.step_target_type === '') {
    stepDetails = (
      <div className="flex flex-col gap-2">
        <Text variant="reg-14">Nuon system step</Text>
      </div>
    )
  }

  return (
    <>
      <div className="flex flex-col gap-4" key={step?.id}>
        <hgroup className="flex gap-4 items-center">
          <YAStatus
            status={step?.status?.status}
            isSkipped={step?.execution_type === 'skipped'}
          />{' '}
          <Text variant="med-18">{sentanceCase(step?.name)}</Text>
        </hgroup>
        {step?.status?.metadata?.reason ? (
          <Notice
            variant={
              step?.status?.status === 'cancelled' ||
              step.execution_type === 'skipped'
                ? 'warn'
                : step?.status?.status === 'error'
                  ? 'error'
                  : 'default'
            }
          >
            {sentanceCase(step?.status?.metadata?.reason as string)}
          </Notice>
        ) : null}
        {stepDetails}
        <Expand
          id={step.id}
          parentClass="border rounded-md"
          headerClass="p-2"
          heading={<Text>View step JSON</Text>}
          expandContent={
            <div className="p-3 border-t">
              <Code variant="preformated">{JSON.stringify(step, null, 2)}</Code>
            </div>
          }
        />
      </div>
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
