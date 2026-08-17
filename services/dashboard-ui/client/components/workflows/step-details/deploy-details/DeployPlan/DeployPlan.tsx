import type { ReactNode } from 'react'
import { HelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { PulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Loading } from '@/components/common/Loading'
import type { TWorkflowStep, TWorkflowStepApprovalType } from '@/types'

type TApprovalType = Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>

type TDiffViewer = Record<TApprovalType, ReactNode>

function getApprovalPlanDiff(step: TWorkflowStep, plan: any): ReactNode {
  const diffs: TDiffViewer = {
    helm_approval: <HelmDiff plan={plan} />,
    kubernetes_manifest_approval: <KubernetesDiff plan={plan?.plan} />,
    terraform_plan: <TerraformDiff plan={plan} />,
    pulumi_plan: <PulumiDiff plan={plan} />,
    app_branch_plan: null,
    install_creation: null,
  }

  return diffs[step?.approval?.type]
}

export interface IDeployPlan {
  step: TWorkflowStep
  plan: any
  isLoading: boolean
  panelId?: string
}

export const DeployPlan = ({ step, plan, isLoading }: IDeployPlan) => {
  return (
    <>
      {isLoading || !plan ? (
        <div className="flex justify-center py-10">
          <Loading variant="large" />
        </div>
      ) : (
        getApprovalPlanDiff(step, plan)
      )}
    </>
  )
}
