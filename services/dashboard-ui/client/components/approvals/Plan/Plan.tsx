import type { ReactNode } from 'react'
import { HelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { PulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Loading } from '@/components/common/Loading'
import type { TWorkflowStep, TWorkflowStepApprovalType } from '@/types'

type TApprovalType = Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>

type TDiffViewer = Record<TApprovalType, ReactNode>

const planLoading = (
  <div className="flex justify-center py-10">
    <Loading variant="large" />
  </div>
)

function getApprovalPlanDiff(step: TWorkflowStep, plan: any): ReactNode {
  const diffs: TDiffViewer = {
    helm_approval: <HelmDiff plan={plan} />,
    kubernetes_manifest_approval: <KubernetesDiff plan={plan} />,
    terraform_plan: <TerraformDiff plan={plan} />,
    pulumi_plan: <PulumiDiff plan={plan} />,
    app_branch_plan: null,
    install_creation: null,
  }
  return diffs[step?.approval?.type]
}

interface IPlanView {
  step: TWorkflowStep
  plan: any
  isLoading: boolean
  error: any
}

export const Plan = ({ step, plan, isLoading, error }: IPlanView) => {
  if (step?.execution_type === 'approval' && !step?.approval) {
    if (!step?.finished) {
      return planLoading
    }
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
      {isLoading && !plan && !error ? (
        planLoading
      ) : !plan && !error ? (
        <EmptyState
          variant="table"
          emptyMessage="The approval plan hasn't been generated yet. Run the workflow to create an approval plan."
          emptyTitle="No plan generated"
        />
      ) : !plan && error ? (
        <EmptyState
          variant="table"
          emptyMessage="Unable to load the approval plan. Try refreshing the page."
          emptyTitle="Loading failed"
        />
      ) : (
        getApprovalPlanDiff(step, plan)
      )}
    </>
  )
}
