'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { ArrowsClockwise, X, Check } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { JsonView } from '@/components/Code'
import { SpinnerSVG } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { approveWorkflowStep } from '@/components/install-actions'
import { DiffEditor, splitYamlDiff } from '@/stratus/components/'
import type { TInstallWorkflowStep } from '@/types'
import { removeSnakeCase } from '@/utils'

interface IApprovalStep {
  approval?: TInstallWorkflowStep['approval']
  buttonText?: string
  headingText?: string
  step: TInstallWorkflowStep
  workflowId: string
  workflowApproveOption?: 'prompt' | 'approve-all'
}

export const ApprovalStep: FC<IApprovalStep> = ({
  approval,
  buttonText = 'Approve changes',
  headingText = 'Approve changes',
  workflowId,
  workflowApproveOption = 'prompt',
  step,
}) => {
  const params = useParams()
  const orgId = params?.['org-id'] as string
  const [isDenyLoading, setIsDenyLoading] = useState(false)
  const [isRetryLoading, setIsRetryLoading] = useState(false)
  const [isApproveLoading, setIsApproveLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()

  const approve = (responseType: 'approve' | 'deny' | 'retry') => {
    setIsKickedOff(true)
    approveWorkflowStep({
      orgId,
      workflowId,
      stepId: step?.id,
      approvalId: approval?.id,
      responseType,
    }).then(({ data, error }) => {
      setIsDenyLoading(false)
      setIsRetryLoading(false)
      setIsApproveLoading(false)
      if (error) {
        setError(error?.error)
        console.error(error)
      } else {
      }
    })
  }

  return (
    <>
      {approval?.response ? (
        approval?.response?.type === 'approve' ? (
          <Notice className="!p-4 w-full" variant="success">
            <Text variant="med-14" className="mb-2">
              Step approved: {removeSnakeCase(approval?.type)}
            </Text>
            <Text isMuted>
              These {removeSnakeCase(approval?.type)} changes have been
              approved.
            </Text>
          </Notice>
        ) : (
          <Notice className="!p-4 w-full" variant="default">
            <Text variant="med-14" className="mb-2">
              Step denied: {removeSnakeCase(approval?.type)}
            </Text>
            <Text isMuted>
              These {removeSnakeCase(approval?.type)} changes have been denied.
            </Text>
          </Notice>
        )
      ) : workflowApproveOption === 'prompt' ? (
        <Notice className="!p-4 w-full" variant="warn">
          <Text variant="med-14" className="mb-2">
            Action needed: {removeSnakeCase(approval?.type)}
          </Text>
          <Text isMuted>
            Approve or deny these changes included in this{' '}
            {removeSnakeCase(approval?.type)}.
          </Text>
        </Notice>
      ) : null}

      <div className="flex flex-col gap-2 !w-full">
        <div className="flex flex-col gap-4 border rounded-md p-2">
          {error ? <Notice>{error}</Notice> : null}
          {approval?.type === 'helm_approval' ? (
            <HelmDiff diff={approval?.contents} />
          ) : (
            <JsonView data={approval?.contents} expanded={2} />
          )}
        </div>
        {approval?.response ||
        workflowApproveOption === 'approve-all' ? null : (
          <div className="mt-4 flex gap-3 justify-end">
            <Button
              onClick={() => {
                setIsDenyLoading(true)
                approve('deny')
              }}
              className="text-sm font-sans flex items-center gap-2 h-[32px]"
              disabled={isKickedOff}
            >
              {isDenyLoading ? (
                <>
                  <SpinnerSVG />
                  Denying plan
                </>
              ) : (
                <>
                  <X />
                  Deny plan
                </>
              )}
            </Button>

            <Button
              onClick={() => {
                setIsRetryLoading(true)
                approve('retry')
              }}
              className="text-sm font-sans flex items-center gap-2 h-[32px]"
              disabled={isKickedOff}
            >
              {isRetryLoading ? (
                <>
                  <SpinnerSVG />
                  Retrying plan
                </>
              ) : (
                <>
                  <ArrowsClockwise />
                  Retry Plan
                </>
              )}
            </Button>
            <Button
              onClick={() => {
                setIsApproveLoading(true)
                approve('approve')
              }}
              className="text-sm font-sans flex items-center gap-2 h-[32px] !px-2"
              disabled={isKickedOff}
              variant="primary"
            >
              {isApproveLoading ? (
                <>
                  <SpinnerSVG />
                  Approving plan
                </>
              ) : (
                <>
                  <Check />
                  Approve plan
                </>
              )}
            </Button>
          </div>
        )}
      </div>
    </>
  )
}

const HelmDiff: FC<{ diff: string }> = ({ diff }) => {
  const splitDiff = splitYamlDiff(diff)
  return <DiffEditor {...splitDiff} language="yaml" />
}
