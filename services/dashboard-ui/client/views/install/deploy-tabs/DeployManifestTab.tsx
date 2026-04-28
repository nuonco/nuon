import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { CodeBlock } from '@/components/common/CodeBlock'
import { EmptyState } from '@/components/common/EmptyState'
import { useOrg } from '@/hooks/use-org'
import { getInstallComponentOutputs } from '@/lib'

export const DeployManifestTab = () => {
  const { componentId, installId } = useParams()
  const { org } = useOrg()

  const { data: outputs } = useQuery({
    queryKey: ['install-component-outputs', org?.id, installId, componentId],
    queryFn: () =>
      getInstallComponentOutputs({
        orgId: org.id,
        installId: installId!,
        componentId: componentId!,
      }),
    enabled: !!org?.id && !!installId && !!componentId,
  })

  const manifest = outputs?.manifest

  if (!manifest) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No manifest"
        emptyMessage="No rendered Kubernetes manifest available for this deploy."
      />
    )
  }

  return (
    <CodeBlock className="!max-h-fit" language="yml">
      {manifest}
    </CodeBlock>
  )
}
