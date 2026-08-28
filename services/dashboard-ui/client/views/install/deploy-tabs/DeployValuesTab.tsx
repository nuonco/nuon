import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { RenderedValues } from '@/components/deploys/RenderedValues'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const DeployValuesTab = () => {
  const { deploy } = useDeploy()
  const { org } = useOrg()
  const { install } = useInstall()

  const planJob =
    deploy?.runner_jobs?.find((j) => j.operation === 'create-apply-plan') ??
    deploy?.runner_jobs?.find((j) => j.operation === 'apply-plan')

  const { data: compositePlan, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-job-plan', org?.id, planJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: planJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!planJob?.id,
  })

  const values = compositePlan?.deploy_plan?.helm?.values

  return (
    <>
      <PageTitle segments={['Deploy values', install?.name]} />
      {isLoading ? (
        <Skeleton height="200px" width="100%" />
      ) : !values || values.length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No values"
          emptyMessage="No Helm values available for this deploy."
        />
      ) : (
        <RenderedValues values={values} />
      )}
    </>
  )
}
