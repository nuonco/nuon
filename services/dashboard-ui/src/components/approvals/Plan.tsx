'use client'

import type { ReactNode } from 'react'
import { HelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep, TWorkflowStepApprovalType } from '@/types'

type TApprovalType = Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>

type TDiffViewer = Record<TApprovalType, ReactNode>

function getApprovalPlanSkeleton(planType: TApprovalType): ReactNode {
  const diffSkeletons: TDiffViewer = {
    helm_approval: <HelmPlanSkeleton />,
    kubernetes_manifest_approval: <KubernetesPlanSkeleton />,
    terraform_plan: <TerraformPlanSkeleton />,
  }

  return diffSkeletons[planType]
}

function getApprovalPlanDiff(step: TWorkflowStep, plan: any): ReactNode {
  const diffs: TDiffViewer = {
    helm_approval: <HelmDiff plan={plan} />,
    kubernetes_manifest_approval: <KubernetesDiff plan={plan} />,
    terraform_plan: <TerraformDiff plan={plan} />,
  }

  return diffs[step?.approval?.type]
}

export const Plan = ({ step }: { step: TWorkflowStep }) => {
  const { plan, isLoading, error } = useQueryApprovalPlan({ step })

  if (step?.execution_type === 'approval' && !step?.approval) {
    return (
      <EmptyState
        variant="table"
        emptyMessage="Unable to load the approval plan changes. Plan would have been discarded if step was retried."
        emptyTitle="No approval plan"
      />
    )
  }

  return (
    <>
      {isLoading ? (
        getApprovalPlanSkeleton(
          (step?.approval?.type as TApprovalType) || 'helm_approval'
        )
      ) : !plan || error ? (
        <EmptyState
          variant="table"
          emptyMessage="Unable to load the approval plan changes."
          emptyTitle="No approval plan"
        />
      ) : (
        getApprovalPlanDiff(step, plan)
      )}
    </>
  )
}

const HelmPlanSkeleton = () => {
  return <Skeleton height="350px" width="100%" />
}

const KubernetesPlanSkeleton = () => {
  return <Skeleton height="350px" width="100%" />
}

const TerraformPlanSkeleton = () => {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton height="350px" width="100%" />
      <Skeleton height="350px" width="100%" />
    </div>
  )
}
