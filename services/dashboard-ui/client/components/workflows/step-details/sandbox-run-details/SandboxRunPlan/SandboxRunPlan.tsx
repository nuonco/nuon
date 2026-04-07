import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Skeleton } from '@/components/common/Skeleton'

export interface ISandboxRunPlan {
  plan: any
  isLoading: boolean
}

export const SandboxRunPlan = ({ plan, isLoading }: ISandboxRunPlan) => {
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
