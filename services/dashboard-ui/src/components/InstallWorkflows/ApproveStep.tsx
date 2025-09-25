'use client'

import classNames from 'classnames'
import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import { ArrowsClockwiseIcon, XIcon, CheckIcon } from '@phosphor-icons/react'
import { approveWorkflowStep } from '@/actions/workflows/approve-workflow-step'
import { Button } from '@/components/Button'
import { JsonView } from '@/components/Code'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Notice, type INotice } from '@/components/Notice'
import { Text } from '@/components/Typography'
//import { approveWorkflowStep } from '@/components/install-actions'
import type { TInstallWorkflowStep } from '@/types'
import { removeSnakeCase } from '@/utils'
import { HelmChangesViewer } from './HelmPlanDiff'
import { TerraformPlanViewer } from './TerraformPlanDiff'
import { KubernetesManifestDiffViewer } from './KubernetesPlanDiff'

const APPROVAL_NOTICE: Record<
  string,
  { title: string; copy: string; variant: INotice['variant'] }
> = {
  approve: {
    title: 'Plan approved',
    copy: 'These changes have been approved and changes will be applied.',
    variant: 'success',
  },
  skip: {
    title: 'Plan skipped',
    copy: 'This changes have been skipped and changes will not be applied.',
    variant: 'warn',
  },
  retry: {
    title: 'Plan being retried',
    copy: 'This plan is being retried, a new plan will be created in the next workflow step.',
    variant: 'info',
  },
  deny: {
    title: 'Plan denied',
    copy: 'This plan was denied and changes will not be applied.',
    variant: 'warn',
  },
  'auto-approve': {
    title: 'Auto approved',
    copy: 'This plan was auto approved and changes will be applied automatically.',
    variant: 'info',
  },
}

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
  const [isPlanLoading, setIsPlanLoading] = useState(true)
  const [plan, setPlan] = useState()
  const [isDenyLoading, setIsDenyLoading] = useState(false)
  const [isRetryLoading, setIsRetryLoading] = useState(false)
  const [isApproveLoading, setIsApproveLoading] = useState(false)
  const [isKickedOff, setIsKickedOff] = useState(false)
  const [error, setError] = useState<string>()

  const fetchPlan = () => {
    fetch(
      `/api/${orgId}/install-workflows/${workflowId}/steps/${step.id}/approvals/${step?.approval?.id}/contents`
    )
      .then((r) => {
        setIsPlanLoading(false)
        return r.json().then((res) => {
          setPlan(res)
        })
      })
      .catch((error) => {
        setError(error?.message || 'Failed to fetch plan')
      })
  }

  useEffect(() => {
    // fetch plan on mount
    fetchPlan()
  }, [])

  useEffect(() => {
    if (approval?.id) {
      fetchPlan()
    }
  }, [approval])

  const approve = (responseType: 'approve' | 'deny' | 'retry') => {
    setIsKickedOff(true)
    approveWorkflowStep({
      approvalId: approval?.id,
      body: { note: '', response_type: responseType },
      orgId,
      workflowId,
      workflowStepId: step?.id,
    }).then(({ data, error }) => {
      setIsDenyLoading(false)
      setIsRetryLoading(false)
      setIsApproveLoading(false)
      if (error) {
        setError(error?.error || `Unable to ${responseType} workflow step plan`)
        console.error(error)
      } else {
      }
    })
  }

  const ApprovalButtons = ({ inBanner = false }: { inBanner?: boolean }) =>
    !approval?.response &&
    workflowApproveOption !== 'approve-all' &&
    step?.status?.status !== 'cancelled' &&
    step?.status?.status !== 'error' &&
    step?.status?.status !== 'auto-skipped' ? (
      <div
        className={classNames('flex items-center gap-4', {
          'self-end ml-auto': !inBanner,
        })}
      >
        <Button
          onClick={() => {
            setIsDenyLoading(true)
            approve('deny')
          }}
          className={classNames(
            'text-sm font-sans flex items-center gap-2 h-[32px] !transition-all',
            {
              '!bg-black/10 dark:!bg-black/50 hover:!bg-black/20 dark:hover:!bg-black/60':
                inBanner,
            }
          )}
          disabled={isKickedOff}
          variant={inBanner ? 'ghost' : 'default'}
        >
          {isDenyLoading ? (
            <>
              <SpinnerSVG />
              Denying plan
            </>
          ) : (
            <>
              <XIcon />
              Deny plan
            </>
          )}
        </Button>

        <Button
          onClick={() => {
            setIsRetryLoading(true)
            approve('retry')
          }}
          className={classNames(
            'text-sm font-sans flex items-center gap-2 h-[32px] !transition-all',
            {
              '!bg-black/10 dark:!bg-black/50 hover:!bg-black/20 dark:hover:!bg-black/60':
                inBanner,
            }
          )}
          disabled={isKickedOff}
          variant={inBanner ? 'ghost' : 'default'}
        >
          {isRetryLoading ? (
            <>
              <SpinnerSVG />
              Retrying plan
            </>
          ) : (
            <>
              <ArrowsClockwiseIcon />
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
              <CheckIcon />
              Approve plan
            </>
          )}
        </Button>
      </div>
    ) : null

  return step.execution_type === 'approval' ? (
    <>
      {approval?.response ? (
        <Notice
          className="!p-4 w-full"
          variant={APPROVAL_NOTICE[approval?.response?.type]?.variant}
        >
          <Text variant="med-14" className="mb-2">
            {APPROVAL_NOTICE[approval?.response?.type]?.title}
          </Text>
          <Text isMuted>{APPROVAL_NOTICE[approval?.response?.type]?.copy}</Text>
        </Notice>
      ) : step?.status?.status === 'auto-skipped' ? (
        <Notice className="!p-4 w-full" variant="info">
          <Text variant="med-14" className="mb-2">
            No changes detected
          </Text>
          <Text isMuted>
            The workflow found no changes to apply. Approval step skipped
            automatically.
          </Text>
        </Notice>
      ) : workflowApproveOption === 'prompt' &&
        step?.status?.status !== 'cancelled' &&
        step?.status?.status !== 'error' &&
        step?.status?.status === 'approval-awaiting' ? (
        <Notice className="!p-4 w-full" variant="warn">
          <div className="flex items-center gap-4">
            <div>
              <Text variant="med-14" className="mb-2">
                Action needed: {removeSnakeCase(approval?.type)}
              </Text>
              <Text isMuted>
                Approve or deny these changes included in this{' '}
                {removeSnakeCase(approval?.type)}.
              </Text>
            </div>
            <ApprovalButtons inBanner />
          </div>
        </Notice>
      ) : null}

      {step?.status?.status === 'success' ||
      step?.status?.status === 'approved' ||
      step?.status?.status === 'approval-awaiting' ||
      step?.status?.status === 'auto-skipped' ? (
        <div className="flex flex-col gap-2 !w-full">
          <div className="flex flex-col gap-4">
            {error ? <Notice>{error}</Notice> : null}
            {isPlanLoading && !plan ? (
              <div className="p-6 mb-2  border rounded-md bg-black/5 dark:bg-white/5">
                <Loading variant="stack" loadingText="Loading plan..." />
              </div>
            ) : approval?.type === 'helm_approval' && plan ? (
              <HelmChangesViewer planData={plan} />
            ) : approval?.type === 'kubernetes_manifest_approval' && plan ? (
              <KubernetesManifestDiffViewer approvalContents={plan} />
            ) : plan ? (
              <TerraformPlanViewer
                plan={plan}
                showNoops={step.status.status === 'auto-skipped'}
              />
            ) : (
              <JsonView data={plan} />
            )}
          </div>
          {workflowApproveOption === 'prompt' &&
          step?.status?.status === 'approval-awaiting' ? (
            <ApprovalButtons />
          ) : null}
        </div>
      ) : null}
    </>
  ) : null
}
