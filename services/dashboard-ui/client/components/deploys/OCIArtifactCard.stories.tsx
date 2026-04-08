export default {
  title: 'Deploys/OCIArtifactCard',
}

import { OCIArtifactCard, OCIArtifactSkeleton } from './OCIArtifactCard'
import { Text } from '@/components/common/Text'

const mockOciArtifact = {
  created_at: '2024-01-15T10:30:00Z',
  tag: 'bldq7fplr1up5atx5zpxotbabm',
  repository: '766121324316.dkr.ecr.us-west-2.amazonaws.com/org-1/app-1',
  size: 15728640,
  digest: 'sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6',
  artifact_type: 'application/vnd.oci.image.manifest.v1+json',
  media_type: 'application/vnd.oci.image.manifest.v1+json',
  os: 'linux',
  architecture: 'amd64',
} as any

export const Default = () => (
  <OCIArtifactCard ociArtifact={mockOciArtifact}>
    <Text variant="subtext">View OCI artifact</Text>
  </OCIArtifactCard>
)

export const Skeleton = () => <OCIArtifactSkeleton />
