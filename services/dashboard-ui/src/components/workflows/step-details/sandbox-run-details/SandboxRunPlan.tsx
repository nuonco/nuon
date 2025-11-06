'use client'

import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Skeleton } from '@/components/common/Skeleton'
import { useQueryApprovalPlan } from '@/hooks/use-query-approval-plan'
import type { TWorkflowStep } from '@/types'

export const SandboxRunPlan = ({ step }: { step: TWorkflowStep }) => {
  const { plan, isLoading } = useQueryApprovalPlan({ step })

  return (
    <>
      {isLoading ? <SandboxRunPlanSkeleton /> : <TerraformDiff plan={plan} />}
    </>
  )
}

export const SandboxRunPlanSkeleton = () => {
  return (
    <div className="flex flex-col gap-6">
      <Skeleton height="350px" width="100%" />
      <Skeleton height="350px" width="100%" />
    </div>
  )
}
