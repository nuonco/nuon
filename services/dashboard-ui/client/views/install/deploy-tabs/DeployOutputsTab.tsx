import { useParams, useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HelmOutputs, HelmOutputsSkeleton } from '@/components/deploys/outputs/HelmOutputs/HelmOutputs'
import { KubernetesRenderedValues } from '@/components/deploys/KubernetesRenderedValues'
import { EmptyState } from '@/components/common/EmptyState'
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
    if (component?.type === 'helm_chart') return <HelmOutputsSkeleton />
    return null
  }

  if (error || !outputs) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No outputs"
        emptyMessage="No outputs available for this component yet."
      />
    )
  }

  if (component?.type === 'helm_chart') {
    return <HelmOutputs createdAt={deploy?.created_at} outputs={outputs} />
  }

  if (component?.type === 'kubernetes_manifest') {
    const values = outputs?.rendered_values
    if (!values) {
      return (
        <EmptyState
          variant="table"
          emptyTitle="No outputs"
          emptyMessage="No Kubernetes values available for this component."
        />
      )
    }
    return <KubernetesRenderedValues values={values} />
  }

  return null
}
