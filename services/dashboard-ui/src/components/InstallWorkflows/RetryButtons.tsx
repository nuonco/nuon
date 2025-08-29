'use client'

import React, { useEffect, useState } from 'react'
import { Button } from '@/components/Button'
import { Notice } from '@/components/Notice'
import { retryWorkflow } from '@/components/install-actions'
import { useOrg } from '@/components/Orgs'
import type { TInstallWorkflowStep } from '@/types'
import { SpinnerSVG } from '../Loading'
import { ArrowClockwiseIcon, SkipForwardIcon } from '@phosphor-icons/react'

export const RetryButtons = ({ step }: { step: TInstallWorkflowStep }) => {
  const { org } = useOrg()
  const [isRetryLoading, setIsRetryLoading] = useState(false)
  const [isSkipLoading, setIsSkipLoading] = useState(false)
  const [isRetryKickedOff, setIsRetryKickedOff] = useState(false)
  const [isSkipKickedOff, setIsSkipKickedOff] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()
  const [shouldRender, setShouldRender] = useState(
    (step?.retryable && step?.status?.status == 'error' && !step?.retried) ||
    step?.status?.status === 'noop'
  )

  useEffect(() => {
    if (step?.retryable && step?.status?.status == 'error' && !step?.retried) {
      setShouldRender(true)
    }
  }, [step])

  const retry = (op: 'retry-step' | 'skip-step') => {
    retryWorkflow({
      orgId: org.id,
      workflowId: step?.install_workflow_id,
      stepId: step?.id,
      op: op,
    }).then(({ data, error }) => {
      switch (op) {
        case 'retry-step':
          setIsRetryKickedOff(true)
          break
        case 'skip-step':
          setIsSkipKickedOff(true)
          break
      }
      if (error) {
        setError(error?.error)
        console.error(error)
      } else {
        setShouldRender(false)
      }
    })
  }

  return (
    <div className="mt-4 flex flex-col gap-4">
      {shouldRender ? (
        <span className="flex justify-between">
          {step?.skippable ? (
            <Button
              onClick={() => {
                setIsKickedOff(true)
                setIsSkipKickedOff(true)
                setIsSkipLoading(true)
                retry('skip-step')
              }}
              variant={'secondary'}
              className="w-fit text-sm"
              disabled={isKickedOff}
            >
              {isSkipLoading ? (
                <>
                  <SpinnerSVG />
                  Continuing
                </>
              ) : (
                <>
                  Skip and continue
                  <SkipForwardIcon />
                </>
              )}
            </Button>
          ) : (<div></div>)}
          <Button
            onClick={() => {
              setIsKickedOff(true)
              setIsRetryKickedOff(true)
              setIsRetryLoading(true)
              retry('retry-step')
            }}
            variant={'primary'}
            className="w-fit text-sm"
            disabled={isKickedOff}
          >
            {isRetryLoading ? (
              <>
                <SpinnerSVG />
                Retrying
              </>
            ) : (
              <>
                Retry
                <ArrowClockwiseIcon />
              </>
            )}
          </Button>
        </span>
      ) : null}
      {error ? <Notice>{error}</Notice> : null}
      {isRetryKickedOff ? (
        <Notice variant="info">Step was discarded and will retry.</Notice>
      ) : null}
      {isSkipKickedOff ? (
        <Notice variant="info">Step was skipped and will continue.</Notice>
      ) : null}
    </div>
  )
}
