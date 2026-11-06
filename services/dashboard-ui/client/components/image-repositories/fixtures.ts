import {
  verifiedECRConnection,
  verifiedGARConnection,
  acrConnection,
} from '@/components/cloud-connections/__stories__/fixtures'
import type { TImageRepository } from '@/types'

export const repositories: TImageRepository[] = [
  {
    id: 'imr-ecr',
    org_id: 'org-mock-001',
    platform: 'aws',
    repository: 'arn:aws:ecr:us-west-2:123456789012:repository/acme/api',
    image_url_prefix: '123456789012.dkr.ecr.us-west-2.amazonaws.com/acme/api',
    cloud_connection_id: verifiedECRConnection.id,
    cloud_connection: verifiedECRConnection,
    created_at: '2026-11-02T07:00:00Z',
    updated_at: '2026-11-02T07:00:00Z',
  },
  {
    id: 'imr-gar',
    org_id: 'org-mock-001',
    platform: 'gcp',
    repository:
      'projects/acme-project/locations/us-central1/repositories/acme-services',
    image_url_prefix: 'us-central1-docker.pkg.dev/acme-project/acme-services',
    cloud_connection_id: verifiedGARConnection.id,
    cloud_connection: verifiedGARConnection,
    created_at: '2026-11-02T07:00:00Z',
    updated_at: '2026-11-02T07:00:00Z',
  },
  {
    id: 'imr-acr',
    org_id: 'org-mock-001',
    platform: 'azure',
    repository: 'acme/api',
    image_url_prefix: 'acmeimages-abc123.azurecr.io/acme/api',
    cloud_connection_id: acrConnection.id,
    cloud_connection: acrConnection,
    created_at: '2026-11-02T07:00:00Z',
    updated_at: '2026-11-02T07:00:00Z',
  },
]
