import { useQuery } from '@tanstack/react-query'
import { RenderedValues } from '@/components/deploys/RenderedValues'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { useDeploy } from '@/hooks/use-deploy'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const DeployValuesTab = () => {
  const { deploy } = useDeploy()
  const { org } = useOrg()

  const planJob = deploy?.runner_jobs?.find(
    (j) => j.operation === 'create-apply-plan'
  ) ?? deploy?.runner_jobs?.find(
    (j) => j.operation === 'apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    queryKey: ['runner-job-plan', org?.id, planJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: planJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!planJob?.id,
  })

  if (isLoading) return <Skeleton height="200px" width="100%" />

  const values = compositePlan?.deploy_plan?.helm?.values

  if (!values || values.length === 0) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No values"
        emptyMessage="No Helm values available for this deploy."
      />
    )
  }

  return <RenderedValues values={values} />
}
