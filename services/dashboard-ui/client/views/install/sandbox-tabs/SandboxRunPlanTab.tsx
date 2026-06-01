import { useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Plan } from '@/components/approvals/Plan'
import { PulumiDiff } from '@/components/approvals/plan-diffs/pulumi/PulumiDiff'
import { TerraformDiff } from '@/components/approvals/plan-diffs/terraform/TerraformDiff'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'
import type { TSandboxRunOutletContext } from './types'

const SandboxRunPlanFallback = () => {
  const { sandboxRun } = useSandboxRun()
  const { org } = useOrg()

  const isPulumi = sandboxRun?.app_sandbox_config?.type === 'pulumi'

  const applyJob = sandboxRun?.runner_jobs?.find(
    (j) => j.operation === 'apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    queryKey: ['runner-job-plan', org?.id, applyJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: applyJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!applyJob?.id,
  })

  if (isLoading) return <Skeleton height="350px" width="100%" />

  const planDisplay = compositePlan?.sandbox_run_plan?.apply_plan_display
  if (!planDisplay) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No plan available"
        emptyMessage="The plan hasn't been generated yet or is not available for this sandbox run."
      />
    )
  }

  const parsed = typeof planDisplay === 'string' ? JSON.parse(planDisplay) : planDisplay

  if (isPulumi) {
    return <PulumiDiff plan={parsed} />
  }

  return <TerraformDiff plan={parsed} />
}

export const SandboxRunPlanTab = () => {
  const { step } = useOutletContext<TSandboxRunOutletContext>()

  if (step?.approval) {
    return <Plan step={step} />
  }

  return <SandboxRunPlanFallback />
}
