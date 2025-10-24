'use client'

import { Card } from '@/components/common/Card'
import { Duration } from '@/components/common/Duration'
import { ID } from '@/components/common/ID'
import { Icon } from '@/components/common/Icon'
import { ContextTooltip } from '@/components/common/ContextTooltip'
import { Skeleton } from '@/components/common/Skeleton'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TDeploy } from '@/types'
import { formatBytes } from '@/utils/string-utils'

export const OCIArtifactCard = ({
  ociArtifact,
}: {
  ociArtifact: TDeploy['oci_artifact']
}) => {
  return (
    <ContextTooltip
      width="w-fit"
      title="OCI artifact"
      items={[
        {
          id: `oci-artifact-date-`,
          title: 'Created',
          subtitle: (
            <Time
              className="!text-nowrap"
              time={ociArtifact?.created_at}
              format="log-datetime"
              variant="label"
              theme="neutral"
            />
          ),
        },
        {
          id: `oci-artifact-id-`,
          title: 'Tag',
          subtitle: ociArtifact?.tag,
        },
        {
          id: `oci-artifact-repo-`,
          title: 'Repository',
          subtitle: ociArtifact?.repository,
        },

        {
          id: `oci-artifact-size-`,
          title: 'Size',
          subtitle: formatBytes(ociArtifact?.size),
        },
        {
          id: `oci-artifact-digest-`,
          title: 'Digest',
          subtitle: (
            <ID variant="label">
              <span className="block truncate max-w-40">
                {ociArtifact?.digest}
              </span>
            </ID>
          ),
        },
        {
          id: `oci-artifact-type-`,
          title: 'Artifact type',
          subtitle: ociArtifact?.artifact_type,
        },
        {
          id: `oci-artifact-media-`,
          title: 'Media type',
          subtitle: (
            <Text className="!text-nowrap" variant="label" theme="neutral">
              {ociArtifact?.media_type}
            </Text>
          ),
        },
        {
          id: `oci-artifact-urls-`,
          title: 'URLs',
          subtitle: ociArtifact?.urls?.toString(),
        },
        {
          id: `oci-artifact-os-`,
          title: 'OS',
          subtitle: ociArtifact?.os,
        },
        {
          id: `oci-artifact-arch-`,
          title: 'Architecture',
          subtitle: ociArtifact?.architecture,
        },
        {
          id: `oci-artifact-variant-`,
          title: 'Variant',
          subtitle: ociArtifact?.variant,
        },
        {
          id: `oci-artifact-os-version-`,
          title: 'OS version',
          subtitle: ociArtifact?.os_version,
        },
        {
          id: `oci-artifact-os-features-`,
          title: 'OS features',
          subtitle: ociArtifact?.os_features?.toString(),
        },
      ]}
    >
      <Card className="!p-2 !flex-row">
        <Text className="flex items-center gap-2" weight="strong">
          OCI artifact <Icon variant="Question" />
        </Text>
      </Card>
    </ContextTooltip>
  )
}

export const OCIArtifactSkeleton = () => {
  return <Skeleton height="42px" width="240px" />
}
