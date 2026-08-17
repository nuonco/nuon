import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Loading } from '@/components/common/Loading'

export interface ISandboxRunPlan {
  plan: any
  isLoading: boolean
}

export const SandboxRunPlan = ({ plan, isLoading }: ISandboxRunPlan) => {
  return (
    <>
      {isLoading ? (
        <div className="flex justify-center py-10">
          <Loading variant="large" />
        </div>
      ) : (
        <TerraformDiff plan={plan} />
      )}
    </>
  )
}
