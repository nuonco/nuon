import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { TerraformRenderedVariables } from '@/components/deploys/TerraformRenderedVariables'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const DeployVariablesTab = () => {
  const { deploy } = useDeploy()
  const { org } = useOrg()
  const { install } = useInstall()

  const planJob = deploy?.runner_jobs?.find(
    (j) => j.operation === 'create-apply-plan'
  ) ?? deploy?.runner_jobs?.find(
    (j) => j.operation === 'apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-job-plan', org?.id, planJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: planJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!planJob?.id,
  })

  const vars = compositePlan?.deploy_plan?.terraform?.vars

  return (
    <>
      <PageTitle segments={['Deploy variables', install?.name]} />
      {isLoading ? (
        <Skeleton height="200px" width="100%" />
      ) : !vars || Object.keys(vars).length === 0 ? (
        <EmptyState
          variant="table"
          emptyTitle="No variables"
          emptyMessage="No Terraform variables available for this deploy."
        />
      ) : (
        <TerraformRenderedVariables values={vars} />
      )}
    </>
  )
}
