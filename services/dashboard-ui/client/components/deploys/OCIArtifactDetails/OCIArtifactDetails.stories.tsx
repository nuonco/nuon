export default {
  title: 'Deploys/OCIArtifactDetails',
}

import { OCIArtifactDetails } from './OCIArtifactDetails'

const mockArtifact = {
  tag: 'bld5llwr0jrd12uqxde8qkb3vm',
  repository: '543968496418.dkr.ecr.us-west-2.amazonaws.com/inl0w4zh66yffrj5vqqbbhrswq',
  digest: 'sha256:7d6a3c8f91470a23ef380320609ee6e69ac68d20bc804f3a1c6065fb56cfa34e',
  size: 1751,
  artifact_type: 'application/vnd.oci.image.manifest.v1+json',
  media_type: 'application/vnd.docker.distribution.manifest.list.v2+json',
  created_at: '2026-04-21T19:35:00Z',
  os: 'linux',
  architecture: 'amd64',
}

const mockArtifactMinimal = {
  tag: 'bld5llwr0jrd12uqxde8qkb3vm',
  repository: '543968496418.dkr.ecr.us-west-2.amazonaws.com/inl0w4zh66yffrj5vqqbbhrswq',
  digest: 'sha256:7d6a3c8f91470a23ef380320609ee6e69ac68d20bc804f3a1c6065fb56cfa34e',
  size: 1751,
  artifact_type: '',
  media_type: 'application/vnd.docker.distribution.manifest.list.v2+json',
  created_at: '2026-04-21T19:35:00Z',
}

export const Default = () => <OCIArtifactDetails artifact={mockArtifact} />

export const Minimal = () => <OCIArtifactDetails artifact={mockArtifactMinimal} />

export const Empty = () => <OCIArtifactDetails artifact={undefined} />
