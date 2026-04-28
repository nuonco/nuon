import { useQuery } from '@tanstack/react-query'
import { TerraformRenderedVariables } from '@/components/deploys/TerraformRenderedVariables'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useDeploy } from '@/hooks/use-deploy'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const DeployVariablesTab = () => {
  const { deploy } = useDeploy()
  const { org } = useOrg()

  const planJob = deploy?.runner_jobs?.find(
    (j) => j.operation === 'create-apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    queryKey: ['runner-job-plan', org?.id, planJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: planJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!planJob?.id,
  })

  if (isLoading) return <Skeleton height="200px" width="100%" />

  const vars = compositePlan?.deploy_plan?.terraform?.vars

  if (!vars || Object.keys(vars).length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No variables"
        emptyMessage="No Terraform variables available for this deploy."
      />
    )
  }

  return <TerraformRenderedVariables values={vars} />
}
