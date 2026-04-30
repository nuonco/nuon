import { useParams } from 'react-router'
import { useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HelmOutputs, HelmOutputsSkeleton } from '@/components/deploys/outputs/HelmOutputs/HelmOutputs'
import { EmptyState } from '@/components/common/EmptyState'
import { Skeleton } from '@/components/common/Skeleton'
import { TerraformOutputs } from '@/components/terraform-outputs/TerraformOutputs'
import { useOrg } from '@/hooks/use-org'
import { useDeploy } from '@/hooks/use-deploy'
import { getInstallComponentOutputs } from '@/lib'
import type { TDeployOutletContext } from './types'

export const DeployOutputsTab = () => {
  const { componentId, installId } = useParams()
  const { component } = useOutletContext<TDeployOutletContext>()
  const { org } = useOrg()
  const { deploy } = useDeploy()

  const { data: outputs, isLoading, error } = useQuery({
    queryKey: ['install-component-outputs', org?.id, installId, componentId],
    queryFn: () =>
      getInstallComponentOutputs({
        orgId: org.id,
        installId: installId!,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!installId && !!componentId,
    retry: false,
  })

  if (isLoading) {
    return component?.type === 'helm_chart'
      ? <HelmOutputsSkeleton />
      : <Skeleton height="400px" width="100%" />
  }

  if (error || !outputs) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No outputs"
        emptyMessage="No outputs available for this component yet."
      />
    )
  }

  if (component?.type === 'helm_chart') {
    return <HelmOutputs createdAt={deploy?.created_at} outputs={outputs} />
  }

  return (
    <TerraformOutputs
      heading={`${component?.name ?? 'Component'} outputs`}
      outputs={outputs}
    />
  )
}
