import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useDeploy } from '@/hooks/use-deploy'
import { formatBytes } from '@/utils/string-utils'

export const DeployArtifactTab = () => {
  const { deploy } = useDeploy()
  const artifact = deploy?.oci_artifact

  if (!artifact) {
    return (
      <EmptyState
        variant="table"
        emptyTitle="No artifact"
        emptyMessage="No OCI artifact available for this deploy."
      />
    )
  }

  return (
    <Card>
      <div className="flex flex-col gap-4 p-4">
        <Text variant="base" weight="strong">OCI artifact</Text>
        <div className="grid grid-cols-2 gap-4">
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Tag</Text>
            <Text variant="subtext" family="mono">{artifact.tag}</Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Repository</Text>
            <Text variant="subtext" family="mono" className="break-all">{artifact.repository}</Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Digest</Text>
            <ID variant="subtext">{artifact.digest}</ID>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Size</Text>
            <Text variant="subtext">{formatBytes(artifact.size)}</Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Artifact type</Text>
            <Text variant="subtext">{artifact.artifact_type}</Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Media type</Text>
            <Text variant="subtext" family="mono">{artifact.media_type}</Text>
          </div>
          <div className="flex flex-col gap-1">
            <Text variant="label" theme="neutral">Created</Text>
            <Time variant="subtext" time={artifact.created_at} format="relative" />
          </div>
          {artifact.os || artifact.architecture ? (
            <div className="flex flex-col gap-1">
              <Text variant="label" theme="neutral">Platform</Text>
              <Text variant="subtext">{[artifact.os, artifact.architecture].filter(Boolean).join('/')}</Text>
            </div>
          ) : null}
        </div>
      </div>
    </Card>
  )
}
