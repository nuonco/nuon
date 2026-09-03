import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { ClickToCopyButton } from '@/components/common/ClickToCopy'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useDeploy } from '@/hooks/use-deploy'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getRunnerJobPlan } from '@/lib'

export const DeployManifestTab = () => {
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

  const manifest = compositePlan?.deploy_plan?.kubernetes_manifest?.manifest

  return (
    <>
      <PageTitle segments={['Deploy manifest', install?.name]} />
      {isLoading ? (
        <Skeleton height="400px" width="100%" />
      ) : !manifest ? (
        <EmptyState
          variant="table"
          emptyTitle="No manifest"
          emptyMessage="No rendered Kubernetes manifest available for this deploy."
        />
      ) : (
        <div className="relative">
          <div className="absolute w-fit top-2 right-2 z-10">
            <ClickToCopyButton textToCopy={manifest} />
          </div>
          <CodeBlock className="!max-h-fit" language="yml">
            {manifest}
          </CodeBlock>
        </div>
      )}
    </>
  )
}
