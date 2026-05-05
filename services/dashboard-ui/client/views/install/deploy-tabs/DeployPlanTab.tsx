import { useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Plan } from '@/components/approvals/Plan'
import { HelmDiff } from '@/components/approvals/plan-diffs/helm/HelmDiff'
import { KubernetesDiff } from '@/components/approvals/plan-diffs/kubernetes/KubernetesDiff'
import { PulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useDeploy } from '@/hooks/use-deploy'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'
import type { TComponentType } from '@/types'
import type { TDeployOutletContext } from './types'

function getDiffForComponentType(componentType: TComponentType, plan: any) {
  switch (componentType) {
    case 'terraform_module':
      return <TerraformDiff plan={plan} />
    case 'helm_chart':
      return <HelmDiff plan={plan} />
    case 'kubernetes_manifest':
      return <KubernetesDiff plan={plan} />
    case 'pulumi':
      return <PulumiDiff plan={plan} />
    default:
      return null
  }
}

const DeployPlanFallback = ({ componentType }: { componentType: TComponentType }) => {
  const { deploy } = useDeploy()
  const { org } = useOrg()

  const applyJob = deploy?.runner_jobs?.find(
    (j) => j.operation === 'apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    queryKey: ['runner-job-plan', org?.id, applyJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: applyJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!applyJob?.id,
  })

  if (isLoading) return <Skeleton height="350px" width="100%" />

  const planDisplay = compositePlan?.deploy_plan?.apply_plan_display
  if (!planDisplay) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan available"
        emptyMessage="The plan hasn't been generated yet or is not available for this deploy."
      />
    )
  }

  const parsed = typeof planDisplay === 'string' ? JSON.parse(planDisplay) : planDisplay
  const diff = getDiffForComponentType(componentType, parsed)

  if (!diff) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan available"
        emptyMessage="Plan view is not supported for this component type."
      />
    )
  }

  return diff
}

export const DeployPlanTab = () => {
  const { step, component } = useOutletContext<TDeployOutletContext>()

  if (step?.approval) {
    return <Plan step={step} />
  }

  return <DeployPlanFallback componentType={component.type!} />
}
