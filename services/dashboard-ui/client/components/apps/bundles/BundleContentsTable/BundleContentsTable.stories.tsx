import type { TAirgapBundleArtifact } from '@/types'
import { BundleContentsTable } from './BundleContentsTable'

export default {
  title: 'Apps/Bundles/BundleContentsTable',
}

const mockArtifacts: TAirgapBundleArtifact[] = [
  {
    id: 'agafx1',
    kind: 'component',
    logical_name: 'api-gateway',
    component_id: 'cmp123',
    digest:
      'sha256:4f1c9a2b6e8d0c3a5b7d9f1e3c5a7b9d1f3e5c7a9b1d3f5e7c9a1b3d5f7e9c1a',
    media_type: 'application/vnd.nuon.component.config.v1+json',
    size: 20480,
  },
  {
    id: 'agafx2',
    kind: 'image',
    logical_name: 'guestbook-ui',
    repository: 'public.ecr.aws/demo/guestbook-ui',
    digest:
      'sha256:9d8c7b6a5f4e3d2c1b0a9f8e7d6c5b4a3f2e1d0c9b8a7f6e5d4c3b2a1f0e9d8c',
    media_type: 'application/vnd.oci.image.manifest.v1+json',
    size: 104857600,
  },
  {
    id: 'agafx5',
    kind: 'sandbox',
    logical_name: 'aws-eks-sandbox',
    digest:
      'sha256:3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d',
    media_type: 'application/vnd.nuon.sandbox.config.v1+json',
    size: 15360,
  },
  {
    id: 'agafx6',
    kind: 'action_step',
    logical_name: 'restart-deployment',
    action_workflow_id: 'acw123',
    digest:
      'sha256:4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e',
    media_type: 'application/vnd.nuon.action.step.v1+json',
    size: 8192,
  },
  {
    id: 'agafx3',
    kind: 'runner_image',
    logical_name: 'nuon-runner',
    repository: 'public.ecr.aws/nuon/runner',
    digest:
      'sha256:1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b',
    media_type: 'application/vnd.oci.image.manifest.v1+json',
    size: 524288000,
  },
  {
    id: 'agafx4',
    kind: 'stack_asset',
    logical_name: 'cloudformation-template',
    digest:
      'sha256:2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b3c',
    media_type: 'application/x-yaml',
    size: 30720,
  },
]

export const Default = () => (
  <BundleContentsTable
    artifacts={mockArtifacts}
    orgId="org123"
    appId="app123"
  />
)

export const Empty = () => <BundleContentsTable artifacts={[]} />
