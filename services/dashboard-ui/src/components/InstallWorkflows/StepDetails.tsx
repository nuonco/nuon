import React, { useState, type FC } from 'react'
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
import { RetryButtons } from './RetryButtons'

export function getStepType(
  step: TInstallWorkflowStep,
  install: TInstall,
  workflowApproveOption: 'prompt' | 'approve-all'
): React.ReactNode {
  let stepDetails = <Loading loadingText="Waiting on step..." variant="page" />

  switch (step.step_target_type) {
    case 'install_sandbox_runs':
      stepDetails = (
        <SandboxStepDetails
          step={step}
          shouldPoll
          workflowApproveOption={workflowApproveOption}
        />
      )
      break

    case 'install_stack_versions':
      stepDetails = (
        <StackStep
          step={step}
          install={install}
          appId={install?.app_id}
          shouldPoll
        />
      )
      break

    case 'install_action_workflow_runs':
      stepDetails = <ActionStepDetails step={step} shouldPoll />
      break

    case 'runners':
      stepDetails = <RunnerStepDetails step={step} shouldPoll />
      break
    case 'install_deploys':
      stepDetails = (
          <DeployStepDetails
            step={step}
            shouldPoll
            workflowApproveOption={workflowApproveOption}
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
            isRetried={step?.retried}
          />{' '}
          <Text variant="med-18">{sentanceCase(step?.name)}</Text>
        </hgroup>
        {step?.status?.metadata?.reason &&
        step?.status?.metadata?.reason !== '' ? (
          <Notice
            variant={
              step?.status?.status === 'cancelled' ||
              step?.status?.status === 'approval-denied' ||
              step.execution_type === 'skipped'
                ? 'warn'
                : step?.status?.status === 'discarded' ||
                    step?.status?.status === 'noop'
                  ? 'info'
                  : step?.status?.status === 'error'
                    ? 'error'
                    : 'default'
            }
            className="!p-4 w-full"
          >
            <Text variant="med-14" className="mb-2">
              {sentanceCase(
                (step?.status?.status_human_description as string) ||
                  'Component deployment failed.'
              )}
            </Text>
            <Text isMuted>
              {sentanceCase(step?.status?.metadata?.reason as string)}
            </Text>
          </Notice>
        ) : null}
        {step?.status?.metadata?.err_step_message ? (
          <Notice variant="warn">
            {sentanceCase(step?.status?.metadata?.err_step_message as string)}
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
        <RetryButtons step={step} />
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
