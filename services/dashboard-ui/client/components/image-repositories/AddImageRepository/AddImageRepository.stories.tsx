import {
  PageStory,
  verifiedECRConnection,
  verifiedGARConnection,
  verifiedACRConnection,
  noop,
} from '@/components/cloud-connections/__stories__/fixtures'
import { AddImageRepository } from './AddImageRepository'

export default { title: 'Features / Image repositories / Add' }

const connections = [
  verifiedECRConnection,
  {
    ...verifiedECRConnection,
    id: 'cc-ecr-alternative',
    name: 'acme-readonly',
    principal: 'arn:aws:iam::123456789012:role/acme-readonly',
  },
  verifiedGARConnection,
  verifiedACRConnection,
]
const props = {
  orgId: 'org-mock-001',
  connections,
  onSave: noop,
  onCreateIdentity: async () => ({
    ...verifiedECRConnection,
    id: 'cc-new',
    name: 'acme-new-images',
    status: 'pending' as const,
  }),
}

export const Empty = () => (
  <PageStory>
    <AddImageRepository {...props} />
  </PageStory>
)
export const ECR = () => (
  <PageStory>
    <AddImageRepository
      {...props}
      initialValues={{
        platform: 'aws',
        repository: 'acme/api',
        connection_id: verifiedECRConnection.id,
      }}
    />
  </PageStory>
)
export const GAR = () => (
  <PageStory>
    <AddImageRepository
      {...props}
      initialValues={{
        platform: 'gcp',
        repository: 'acme-services',
        connection_id: verifiedGARConnection.id,
      }}
    />
  </PageStory>
)
export const ACR = () => (
  <PageStory>
    <AddImageRepository
      {...props}
      initialValues={{
        platform: 'azure',
        repository: 'acme/api',
        connection_id: verifiedACRConnection.id,
      }}
    />
  </PageStory>
)
export const NoIdentities = () => (
  <PageStory>
    <AddImageRepository {...props} connections={[]} />
  </PageStory>
)
export const NewIdentity = () => (
  <PageStory>
    <AddImageRepository
      {...props}
      initialValues={{
        platform: 'aws',
        repository: 'acme/api',
        connection_id: '',
        identity_mode: 'new',
      }}
    />
  </PageStory>
)
export const Duplicate = () => (
  <PageStory>
    <AddImageRepository
      {...props}
      initialValues={{
        platform: 'aws',
        repository: 'acme/api',
        connection_id: verifiedECRConnection.id,
      }}
      error={{
        error: 'Repository already registered',
        description: 'This repository already has a saved pull identity.',
        user_error: true,
      }}
    />
  </PageStory>
)
