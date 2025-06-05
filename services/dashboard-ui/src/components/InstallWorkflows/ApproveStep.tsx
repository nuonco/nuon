'use client'

import { useParams } from 'next/navigation'
import React, { type FC, useState } from 'react'
import { X, Check } from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { SpinnerSVG } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { CodeViewer, JsonView } from '@/components/Code'
import { Text } from '@/components/Typography'
import { approveWorkflowStep } from '@/components/install-actions'
import { DiffEditor } from '@/stratus/components/common/Code'
import type { TInstallWorkflowStep } from '@/types'
import { removeSnakeCase } from '@/utils'

interface IApprovalStep {
  approval?: TInstallWorkflowStep['approval']
  buttonText?: string
  headingText?: string
  step: TInstallWorkflowStep
  workflowId: string
}

export const ApprovalStep: FC<IApprovalStep> = ({
  approval,
  buttonText = 'Approve changes',
  headingText = 'Approve changes',
  step,
}) => {
  const params = useParams()
  const orgId = params?.['org-id'] as string
  const workflowId = params?.['workflow-id'] as string
  const [isDenyLoading, setIsDenyLoading] = useState(false)
  const [isApproveLoading, setIsApproveLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()

  const approve = (responseType: 'approve' | 'deny') => {
    setIsKickedOff(true)
    approveWorkflowStep({
      orgId,
      workflowId,
      stepId: step?.id,
      approvalId: approval?.id,
      responseType,
    }).then(({ data, error }) => {
      setIsDenyLoading(false)
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
      <Notice className="!p-4 w-full" variant="warn">
        <Text variant="med-14" className="mb-2">
          Action needed: {removeSnakeCase(approval?.type)}
        </Text>
        <div className="flex flex-col gap-2 !w-full">
          <Text isMuted>
            Approve or deny these changes included in this{' '}
            {removeSnakeCase(approval?.type)}.
          </Text>
          <div className="flex flex-col gap-4">
            {error ? <Notice>{error}</Notice> : null}
            {approval?.type === 'helm_approval' ? (
              <div className="rounded-md overflow-hidden border bg-cool-grey-50 dark:bg-dark-grey-200">
                <DiffEditor diff={approval?.contents} />
              </div>
            ) : (
              <JsonView expanded={2} data={approval?.contents} />
            )}
          </div>
          <div className="mt-4 flex gap-3 justify-end">
            <Button
              onClick={() => {
                setIsDenyLoading(true)
                approve('deny')
              }}
              className="text-sm font-sans flex items-center gap-2 hover:!bg-cool-grey-100 hover:dark:!bg-dark-grey-400"
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
                setIsApproveLoading(true)
                approve('approve')
              }}
              className="text-sm font-sans flex items-center gap-2 hover:!bg-cool-grey-100 hover:dark:!bg-dark-grey-400"
              disabled={isKickedOff}
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
        </div>
      </Notice>
    </>
  )
}
