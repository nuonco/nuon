// @ts-nocheck

'use client'

import classNames from 'classnames'
import {
  useSearchParams,
  useRouter,
  usePathname,
  useParams,
} from 'next/navigation'
import React, { type FC, useEffect, useRef, useState } from 'react'
import { Button } from '@/components/Button'
import { Badge } from '@/components/Badge'
import { Empty } from '@/components/Empty'
import { Section } from '@/components/Card'
import { Duration } from '@/components/Time'
import { Text } from '@/components/Typography'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import type { IPollingProps } from "@/hooks/use-polling"
import type {
  TInstallWorkflow,
  TInstallWorkflowStep,
  TWorkflowStep,
  TInstall,
} from '@/types'
import { removeSnakeCase, sentanceCase } from '@/utils'
import { YAStatus } from './InstallWorkflowHistory'
import { StepDetails, getStepType } from './StepDetails'

export interface IPollStepDetails extends IPollingProps {
  step: TInstallWorkflowStep
  workflowApproveOption?: 'prompt' | 'approve-all'
}

interface IInstallWorkflowSteps {
  installWorkflow: TInstallWorkflow
}

export const InstallWorkflowSteps: FC<IInstallWorkflowSteps> = ({
  installWorkflow,
}) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const path = usePathname()
  const router = useRouter()
  const params = useParams()
  const searchParams = useSearchParams()
  const queryTargetId = searchParams.get('target')
  const orgId = org.id
  const workflowSteps = installWorkflow?.steps?.filter(
    (s) => s?.execution_type !== 'hidden'
  )
  const [stepCount, setStepCount] = useState(workflowSteps?.length)
  const [activeStep, setActiveStep] = useState(
    workflowSteps?.find((s) => s?.id === queryTargetId) ||
      workflowSteps?.find(
        (s) =>
          s?.status?.status === 'in-progress' ||
          s?.status?.status === 'approval-awaiting' ||
          s?.status?.status === 'error'
      ) ||
      installWorkflow?.finished
      ? workflowSteps?.find((s) => s?.status?.status === 'error')
      : workflowSteps?.at(-1) ||
          workflowSteps?.find((s) => s?.step_target_type !== '')
  )
  const scrollableRef = useRef(null)
  const buttonRefs = useRef([])
  const buttonOffset = installWorkflow?.finished
    ? 300
    : installWorkflow?.approval_option === 'approve-all'
      ? 325
      : 375
  const [isManualControl, setManualControl] = useState(false)

  useEffect(() => {
    if (!isManualControl) {
      if (workflowSteps?.some((s) => s?.status?.status === 'in-progress')) {
        if (
          activeStep?.id !==
          workflowSteps?.find((s) => s?.status?.status === 'in-progress').id
        ) {
          setActiveStep(
            workflowSteps?.find((s) => s?.status?.status === 'in-progress')
          )
        }
      } else if (!activeStep) {
        setActiveStep(workflowSteps?.at(0))
      }
    } else {
      if (stepCount < workflowSteps?.length) {
        setActiveStep(
          workflowSteps?.find((s) => s?.status?.status === 'in-progress') ||
            workflowSteps?.find((s) => activeStep?.id === s?.id)
        )
        setStepCount(workflowSteps.length)
        setManualControl(false)
      }
    }
  }, [installWorkflow])

  useEffect(() => {
    const activeIndex = workflowSteps?.findIndex((s) => s.id === activeStep?.id)

    if (buttonRefs.current[activeIndex] && !isManualControl) {
      const button = buttonRefs.current[activeIndex]
      const container = scrollableRef.current
      const buttonTop = button.offsetTop
      const newScrollTop = buttonTop - buttonOffset

      container.scrollTo({
        top: newScrollTop,
        behavior: 'smooth',
      })
    }
  }, [activeStep])

  return (
    <div className="grid grid-cols-1 md:grid-cols-12 flex-auto divide-x h-full">
      <div className="md:col-span-4 overflow-auto" ref={scrollableRef}>
        <Section
          className="flex flex-col gap-2"
          childrenClassName="flex flex-col gap-4"
        >
          {workflowSteps?.length ? (
            <div className="flex flex-col gap-2 workflow-steps">
              {(() => {
                const steps =
                  workflowSteps?.filter(
                    (step) => step?.execution_type !== 'hidden'
                  ) || []
                const groupedSteps = steps.reduce((acc, step) => {
                  const key = step.group_idx
                  if (!acc[key]) {
                    acc[key] = []
                  }
                  acc[key].push(step)
                  return acc
                }, {})

                const sortedGroups = Object.entries(groupedSteps)
                  .sort(([, a], [, b]) => a[0].group_idx - b[0].group_idx)
                  .map(([groupId, groupSteps]) =>
                    groupSteps.sort(
                      (a, b) => a.group_retry_idx - b.group_retry_idx
                    )
                  )
                return sortedGroups.map((groupSteps, groupIndex) => (
                  <React.Fragment key={`group-${groupIndex}`}>
                    {groupSteps.map((step, stepIndex) => {
                      const globalIndex = steps.findIndex(
                        (s) => s.id === step.id
                      )
                      const isLastInRetryGroup =
                        stepIndex === groupSteps.length - 1
                      return (
                        <React.Fragment key={step?.id}>
                          {step?.status?.status === 'pending' ? (
                            <div
                              ref={(el) =>
                                (buttonRefs.current[globalIndex] = el)
                              }
                              className={classNames(
                                'p-2 rounded-md !text-cool-grey-600 dark:!text-cool-grey-500 history-event w-full',
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
                                stepNumber={globalIndex + 1}
                                isSkipped={step?.execution_type === 'skipped'}
                                isRetried={step?.retried}
                              />
                            </div>
                          ) : (
                            <Button
                              ref={(el) =>
                                (buttonRefs.current[globalIndex] = el)
                              }
                              className={classNames(
                                'text-left border-none !p-2 history-event w-full',
                                {
                                  '!bg-black/5 dark:!bg-white/5 !text-cool-grey-950 dark:!text-cool-grey-50':
                                    activeStep?.id === step?.id,
                                  '!bg-transparent hover:!bg-black/5 focus:!bg-black/5 active:!bg-black/10 dark:hover:!bg-white/5 dark:focus:!bg-white/5 dark:active:!bg-white/10':
                                    activeStep?.id !== step?.id,
                                }
                              )}
                              onClick={() => {
                                if (!isManualControl) setManualControl(true)
                                if (step?.step_target_id) {
                                  router.push(
                                    `${path}?${new URLSearchParams({ target: step?.id }).toString()}`
                                  )
                                }
                                setActiveStep(step)

                                if (
                                  buttonRefs.current[globalIndex] &&
                                  scrollableRef.current
                                ) {
                                  const button = buttonRefs.current[globalIndex]
                                  const container = scrollableRef.current
                                  const buttonTop = button.offsetTop
                                  const newScrollTop = buttonTop - buttonOffset

                                  container.scrollTo({
                                    top: newScrollTop,
                                    behavior: 'smooth',
                                  })
                                }
                              }}
                            >
                              <InstallWorkflowStepTitle
                                executionTime={step?.execution_time}
                                name={step?.name}
                                status={step?.status}
                                stepNumber={globalIndex + 1}
                                isSkipped={step?.execution_type === 'skipped'}
                                isRetried={step?.retried}
                              />
                            </Button>
                          )}
                          {!isLastInRetryGroup &&
                            step.group_retry_idx <
                              groupSteps[stepIndex + 1]?.group_retry_idx && (
                              <hr className="border-cool-grey-200 dark:border-dark-grey-600 ml-10 mt-2" />
                            )}
                        </React.Fragment>
                      )
                    })}
                    {groupIndex < sortedGroups.length - 1 && (
                      <hr className="border-cool-grey-200 dark:border-dark-grey-600 ml-10 mt-2" />
                    )}
                  </React.Fragment>
                ))
              })()}
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
              activeStepIndex={workflowSteps?.findIndex(
                (s) => s?.id === activeStep?.id
              )}
            >
              {workflowSteps?.map((step) =>
                getStepType(step, installWorkflow?.approval_option)
              )}
            </StepDetails>
          </Section>
        ) : (
          <Section>
            <Empty
              emptyTitle="Waiting on steps"
              emptyMessage="Waiting on workflow steps details."
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
  isSkipped?: boolean
  isRetried?: boolean
  name: string
  status: TInstallWorkflowStep['status']
  stepNumber: number
}> = ({
  executionTime,
  isSkipped = false,
  isRetried = false,
  name,
  status,
  stepNumber,
}) => {
  return (
    <span className="flex gap-2 items-start justify-start w-full">
      <YAStatus
        status={status?.status}
        isSkipped={isSkipped}
        isRetried={isRetried}
      />
      <span className="flex flex-col w-full max-w-full overflow-hidden gap-1">
        <Text variant="med-12">
          <span className="truncate">{sentanceCase(name)}</span>
        </Text>
        <Text
          className="!text-cool-grey-600 dark:!text-cool-grey-500 w-full justify-between"
          variant="reg-12"
        >
          Step {stepNumber} {isRetried ? `retry initiated by the user` : null}
          {isSkipped && status.status === 'success' ? (
            <Badge theme="info" isCompact>
              Skipped
            </Badge>
          ) : status?.status === 'cancelled' ? (
            <Badge theme="warn" isCompact>
              Cancelled
            </Badge>
          ) : status.status === 'error' ? (
            <Badge theme="error" isCompact>
              Failed
            </Badge>
          ) : status.status === 'not-attempted' ? (
            <Badge isCompact>Not attempted</Badge>
          ) : status.status === 'approval-awaiting' ? (
            <Badge isCompact theme="warn">
              Awaiting approval
            </Badge>
          ) : status.status === 'approved' ? (
            <Badge isCompact theme="success">
              Plan approved
            </Badge>
          ) : status.status === 'approval-denied' ? (
            <Badge isCompact theme="warn">
              Approval denied
            </Badge>
          ) : status?.status === 'success' ? (
            <span className="flex gap-1">
              {getFinishedText(status)} in
              <Duration nanoseconds={executionTime} />
            </span>
          ) : status?.status === 'user-skipped' ? (
            <Badge isCompact>Skipped</Badge>
          ) : status?.status === 'auto-skipped' ? (
            <Badge isCompact theme="info">
              Noop
            </Badge>
          ) : status?.status === 'discarded' ? (
            <Badge isCompact>Discarded</Badge>
          ) : null}
        </Text>
      </span>
    </span>
  )
}

function getFinishedText(
  status: TInstallWorkflowStep['status'],
  isSkipped = false
): string {
  let text: string
  switch (status?.status) {
    case 'cancelled':
      text = 'Cancelled'
      break
    case 'error':
      text = 'Failed'
      break
    case 'success':
      text = 'Completed'
      break
    default:
      text = 'Finished'
  }

  if (isSkipped || status === 'auto-approved') {
    text = 'Skipped'
  }

  return text
}
