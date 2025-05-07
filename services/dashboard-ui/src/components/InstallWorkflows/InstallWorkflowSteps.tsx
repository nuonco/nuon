'use client'

import classNames from 'classnames'
import { useSearchParams, useRouter, usePathname } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { Button } from '@/components/Button'
import { Empty } from '@/components/Empty'
import { Notice } from '@/components/Notice'
import { Section } from '@/components/Card'
import { Text } from '@/components/Typography'
import type { TInstallWorkflow, TInstallWorkflowStep, TInstall } from '@/types'
import { removeSnakeCase, sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'
import { StepDetails, getStepType } from './StepDetails'

export interface IPollStepDetails {
  pollDuration?: number
  shouldPoll?: boolean
  step: TInstallWorkflowStep
}

interface IInstallWorkflowSteps {
  install: TInstall
  installWorkflow: TInstallWorkflow
  orgId: string
}

export const InstallWorkflowSteps: FC<IInstallWorkflowSteps> = ({
  install,
  installWorkflow,
}) => {
  const path = usePathname()
  const router = useRouter()
  const searchParams = useSearchParams()
  const queryTargetId = searchParams.get('target')
  const [activeStep, setActiveStep] = useState(
    installWorkflow?.steps.find((s) => s?.step_target_id === queryTargetId) ||
      installWorkflow?.steps?.find(
        (s) => s?.status?.status === 'in-progress'
      ) ||
      installWorkflow?.steps?.find((s) => s?.step_target_type !== '')
  )
  const [isManualControl, setManualControl] = useState(false)

  useEffect(() => {
    if (!isManualControl) {
      if (
        installWorkflow?.steps?.some((s) => s?.status?.status === 'in-progress')
      ) {
        if (
          activeStep?.id !==
          installWorkflow?.steps?.find(
            (s) => s?.status?.status === 'in-progress'
          ).id
        ) {
          setActiveStep(
            installWorkflow?.steps?.find(
              (s) => s?.status?.status === 'in-progress'
            )
          )
        }
      } else if (!activeStep) {
        setActiveStep(installWorkflow?.steps?.at(0))
      }
    }
  }, [installWorkflow])

  return (
    <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x h-full">
      <div className="md:col-span-4">
        <Section
          heading={`${removeSnakeCase(sentanceCase(installWorkflow?.type))} plan`}
          className="flex flex-col gap-2"
          childrenClassName="flex flex-col gap-4"
        >
          {installWorkflow?.steps?.length ? (
            <div className="flex flex-col gap-2 workflow-steps">
              {installWorkflow?.steps?.map((step, i) => {
                return step?.step_target_type === '' ? (
                  <div
                    key={step?.id}
                    className={classNames(
                      'p-2 rounded-md !text-cool-grey-600 dark:!text-cool-grey-500 history-event',
                      {
                        '!bg-black/5 dark:!bg-white/5 !text-cool-grey-950 dark:!text-cool-grey-50':
                          activeStep?.id === step?.id,
                      }
                    )}
                  >
                    <InstallWorkflowStepTitle
                      executionTime={step?.execution_time}
                      name={step?.name}
                      status={step?.status}
                      stepNumber={i + 1}
                    />
                  </div>
                ) : (
                  <Button
                    className={classNames(
                      'text-left border-none !p-2 history-event',
                      {
                        '!bg-black/5 dark:!bg-white/5 !text-cool-grey-950 dark:!text-cool-grey-50':
                          activeStep?.id === step?.id,
                        '!bg-transparent hover:!bg-black/5 focus:!bg-black/5 active:!bg-black/10 dark:hover:!bg-white/5 dark:focus:!bg-white/5 dark:active:!bg-white/10':
                          activeStep?.id !== step?.id,
                      }
                    )}
                    key={step?.id}
                    onClick={() => {
                      if (!isManualControl) setManualControl(true)
                      router.push(
                        `${path}?${new URLSearchParams({ target: step?.step_target_id }).toString()}`
                      )
                      setActiveStep(step)
                    }}
                  >
                    <InstallWorkflowStepTitle
                      executionTime={step?.execution_time}
                      name={step?.name}
                      status={step?.status}
                      stepNumber={i + 1}
                    />
                  </Button>
                )
              })}
            </div>
          ) : (
            <Empty
              emptyTitle="Waiting on steps"
              emptyMessage="Waiting on update steps to generate."
              variant="history"
            />
          )}
        </Section>
      </div>

      <div className="md:col-span-8">
        {activeStep ? (
          <Section>
            <StepDetails
              activeStepIndex={installWorkflow?.steps?.findIndex(
                (s) => s?.id === activeStep?.id
              )}
            >
              {installWorkflow?.steps?.map((step) => (
                <div className="flex flex-col gap-4" key={step?.id}>
                  <hgroup className="flex gap-4 items-center">
                    <YAStatus status={step?.status?.status} />{' '}
                    <Text variant="med-18">{sentanceCase(step?.name)}</Text>
                  </hgroup>
                  {step?.status?.metadata?.reason ? (
                    <Notice variant="warn">
                      {sentanceCase(step?.status?.metadata?.reason as string)}
                    </Notice>
                  ) : null}
                  {getStepType(step, install)}
                </div>
              ))}
            </StepDetails>
          </Section>
        ) : (
          <Section>
            <Empty
              emptyTitle="Waiting on steps"
              emptyMessage="Waiting on update steps to generate."
              variant="history"
            />
          </Section>
        )}
      </div>
    </div>
  )
}

const InstallWorkflowStepTitle: FC<{
  executionTime: number
  name: string
  status: TInstallWorkflowStep['status']
  stepNumber: number
}> = ({ name, status, stepNumber }) => {
  return (
    <span className="flex gap-2 items-start justify-start">
      <YAStatus status={status?.status} />
      <span>
        <Text
          className="!text-cool-grey-600 dark:!text-cool-grey-500"
          variant="reg-12"
        >
          Step {stepNumber}
        </Text>

        <Text variant="med-12">{sentanceCase(name)}</Text>
      </span>
    </span>
  )
}
