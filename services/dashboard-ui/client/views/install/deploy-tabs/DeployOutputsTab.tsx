import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HelmOutputs, HelmOutputsSkeleton } from '@/components/deploys/outputs/HelmOutputs/HelmOutputs'
import { EmptyState } from '@/components/common/EmptyState'
import { useOrg } from '@/hooks/use-org'
import { useDeploy } from '@/hooks/use-deploy'
import { getInstallComponentOutputs } from '@/lib'

export const DeployOutputsTab = () => {
  const { componentId, installId } = useParams()
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

  if (isLoading) return <HelmOutputsSkeleton />

  if (error || !outputs) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No outputs"
        emptyMessage="No outputs available for this component yet."
      />
    )
  }

  return <HelmOutputs createdAt={deploy?.created_at} outputs={outputs} />
}
