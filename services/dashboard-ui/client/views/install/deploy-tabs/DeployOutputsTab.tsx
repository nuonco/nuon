import { useParams, useOutletContext } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { HelmOutputs, HelmOutputsSkeleton } from '@/components/deploys/outputs/HelmOutputs/HelmOutputs'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { Text } from '@/components/common/Text'
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
    const diffs = outputs?.diff
    if (!diffs || !Array.isArray(diffs) || diffs.length === 0) {
      return (
        <EmptyState
          variant="table"
          emptyTitle="No outputs"
          emptyMessage="No resource diffs available for this component."
        />
      )
    }
    return (
      <div className="flex flex-col gap-4">
        {diffs.map((d: any, i: number) => (
          <div key={d.name ?? i} className="flex flex-col gap-1">
            {d.name ? (
              <Text variant="subtext" weight="strong">
                {d.name}
              </Text>
            ) : null}
            <CodeBlock className="!max-h-fit" language="diff">
              {d.diff ?? JSON.stringify(d, null, 2)}
            </CodeBlock>
          </div>
        ))}
      </div>
    )
  }

  return null
}
