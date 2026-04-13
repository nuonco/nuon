import type { ReactNode } from 'react'
import { HelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { PulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { Skeleton } from '@/components/common/Skeleton'
import type { TWorkflowStep, TWorkflowStepApprovalType } from '@/types'

type TApprovalType = Exclude<TWorkflowStepApprovalType, 'approve-all' | 'noop'>

type TDiffViewer = Record<TApprovalType, ReactNode>

function getApprovalPlanSkeleton(planType: TApprovalType): ReactNode {
  const diffSkeletons: TDiffViewer = {
    helm_approval: <HelmPlanSkeleton />,
    kubernetes_manifest_approval: <KubernetesPlanSkeleton />,
    terraform_plan: <TerraformPlanSkeleton />,
    pulumi_plan: <Skeleton height="350px" width="100%" />,
  }

  return diffSkeletons[planType]
}

function getApprovalPlanDiff(step: TWorkflowStep, plan: any): ReactNode {
  const diffs: TDiffViewer = {
    helm_approval: <HelmDiff plan={plan} />,
    kubernetes_manifest_approval: <KubernetesDiff plan={plan?.plan} />,
    terraform_plan: <TerraformDiff plan={plan} />,
    pulumi_plan: <PulumiDiff plan={plan} />,
  }

  return diffs[step?.approval?.type]
}

export interface IDeployPlan {
  step: TWorkflowStep
  plan: any
  isLoading: boolean
  panelId?: string
}

export const DeployPlan = ({
  step,
  plan,
  isLoading,
}: IDeployPlan) => {
  return (
    <>
      {isLoading || !plan
        ? getApprovalPlanSkeleton(step?.approval?.type as TApprovalType)
        : getApprovalPlanDiff(step, plan)}
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
