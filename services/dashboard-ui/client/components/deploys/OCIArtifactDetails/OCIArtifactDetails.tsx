import { Card } from '@/components/common/Card'
import { ClickToCopy } from '@/components/common/ClickToCopy'
import { EmptyState } from '@/components/common/EmptyState'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TDeploy } from '@/types'
import { formatBytes } from '@/utils/string-utils'

export interface IOCIArtifactDetails {
  artifact: TDeploy['oci_artifact']
}

export const OCIArtifactDetails = ({ artifact }: IOCIArtifactDetails) => {
  if (!artifact) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No artifact"
        emptyMessage="No OCI artifact available for this deploy."
      />
    )
  }

  return (
    <Card>
      <Text variant="base" weight="strong">OCI artifact</Text>

      <div className="grid grid-cols-2 lg:grid-cols-3 gap-6">
        <LabeledValue label="Tag">
          <ClickToCopy>
            <Text variant="subtext" family="mono">{artifact.tag}</Text>
          </ClickToCopy>
        </LabeledValue>

        <LabeledValue label="Repository" className="lg:col-span-2">
          <ClickToCopy>
            <Text variant="subtext" family="mono" className="break-all">{artifact.repository}</Text>
          </ClickToCopy>
        </LabeledValue>

        <LabeledValue label="Digest">
          <ClickToCopy>
            <Text variant="subtext" family="mono" className="break-all">{artifact.digest}</Text>
          </ClickToCopy>
        </LabeledValue>

        <LabeledValue label="Size">
          <Text variant="subtext">{formatBytes(artifact.size)}</Text>
        </LabeledValue>

        <LabeledValue label="Created">
          <Time variant="subtext" time={artifact.created_at} format="relative" />
        </LabeledValue>

        <LabeledValue label="Artifact type">
          <Text variant="subtext">{artifact.artifact_type}</Text>
        </LabeledValue>

        <LabeledValue label="Media type">
          <Text variant="subtext" family="mono">{artifact.media_type}</Text>
        </LabeledValue>

        {artifact.os || artifact.architecture ? (
          <LabeledValue label="Platform">
            <Text variant="subtext">
              {[artifact.os, artifact.architecture].filter(Boolean).join('/')}
            </Text>
          </LabeledValue>
        ) : null}
      </div>
    </Card>
  )
}
