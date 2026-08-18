import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { TerraformRenderedVariables } from '@/components/deploys/TerraformRenderedVariables'
import { TerraformRenderedVariablesFiles } from '@/components/deploys/TerraformRenderedVariablesFiles'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { useSandboxRun } from '@/hooks/use-sandbox-run'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const SandboxRunVariablesTab = () => {
  const { sandboxRun } = useSandboxRun()
  const { org } = useOrg()

  const planJob = sandboxRun?.runner_jobs?.find(
    (j) => j.operation === 'create-apply-plan'
  ) ?? sandboxRun?.runner_jobs?.find(
    (j) => j.operation === 'apply-plan'
  )

  const { data: compositePlan, isLoading } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['runner-job-plan', org?.id, planJob?.id],
    queryFn: () =>
      getRunnerJobPlan({ runnerJobId: planJob!.id, orgId: org.id }),
    enabled: !!org?.id && !!planJob?.id,
  })

  if (isLoading) return <Skeleton height="200px" width="100%" />

  const vars = compositePlan?.sandbox_run_plan?.vars
  const varsFiles = compositePlan?.sandbox_run_plan?.vars_files as
    | string[]
    | undefined

  const hasVars = !!vars && Object.keys(vars).length > 0
  const hasVarsFiles = !!varsFiles && varsFiles.length > 0

  if (!hasVars && !hasVarsFiles) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No variables"
        emptyMessage="No Terraform variables available for this sandbox run."
      />
    )
  }

  return (
    <div className="flex flex-col gap-6">
      {hasVars && (
        <div className="flex flex-col gap-2">
          <Text weight="strong">Variables</Text>
          <TerraformRenderedVariables values={vars} />
        </div>
      )}
      {hasVarsFiles && (
        <div className="flex flex-col gap-2">
          <Text weight="strong">Variable files</Text>
          <TerraformRenderedVariablesFiles files={varsFiles} />
        </div>
      )}
    </div>
  )
}
