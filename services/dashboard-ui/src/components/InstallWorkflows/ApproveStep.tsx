'use client'

import classNames from 'classnames'
import { useParams } from 'next/navigation'
import React, { type FC, useEffect, useState } from 'react'
import {
  ArrowsClockwise,
  X,
  Check,
  ArrowRightIcon,
  ArrowsOutSimpleIcon,
  CaretDownIcon,
  CaretRightIcon,
} from '@phosphor-icons/react'
import { Button } from '@/components/Button'
import { JsonView } from '@/components/Code'
import { SpinnerSVG, Loading } from '@/components/Loading'
import { Notice } from '@/components/Notice'
import { Text } from '@/components/Typography'
import { approveWorkflowStep } from '@/components/install-actions'
import { DiffEditor, splitYamlDiff } from '@/stratus/components/'
import type { TInstallWorkflowStep } from '@/types'
import { removeSnakeCase } from '@/utils'
import { HelmChangesViewer } from './HelmPlanDiff'
import { TerraformPlanViewer } from './TerraformPlanDiff'
import { KubernetesManifestDiffViewer } from './KubernetesPlanDiff'

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

  useEffect(() => {
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
  }, [])

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

  const ApprovalButtons = ({ inBanner = false }: { inBanner?: boolean }) =>
    approval?.response ||
    workflowApproveOption === 'approve-all' ||
    step?.status?.status === 'cancelled' ? null : (
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
    )

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
        ) : approval?.response?.type === 'deny' ? (
          <Notice className="!p-4 w-full" variant="default">
            <Text variant="med-14" className="mb-2">
              Step denied: {removeSnakeCase(approval?.type)}
            </Text>
            <Text isMuted>
              These {removeSnakeCase(approval?.type)} changes have been denied.
            </Text>
          </Notice>
        ) : (
          <Notice className="!p-4 w-full" variant="default">
            <Text variant="med-14" className="mb-2">
              Step retry: {removeSnakeCase(approval?.type)}
            </Text>
            <Text isMuted>
              These {removeSnakeCase(approval?.type)} changes have been retried.
            </Text>
          </Notice>
        )
      ) : workflowApproveOption === 'prompt' &&
        step?.status?.status !== 'cancelled' && step?.status?.status !== 'auto-skipped' ? (
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
            <TerraformPlanViewer plan={plan} />
          ) : (
            <JsonView data={plan} />
          )}
        </div>
        { step?.status?.status !== 'cancelled' && step?.status?.status !== 'auto-skipped' ? (
          <ApprovalButtons />
        ) : null}
      </div>
    </>
  )
}

const HelmDiff: FC<{ diff: string }> = ({ diff }) => {
  const splitDiff = splitYamlDiff(diff)
  return <DiffEditor {...splitDiff} language="yaml" />
}
