import {
  PageStory,
  initialECRConnection,
  initialGARConnection,
  initialACRConnection,
  verifiedECRConnection,
  verifiedGARConnection,
  verifiedACRConnection,
  noop,
} from '@/components/cloud-connections/__stories__/fixtures'
import { repositories } from '../fixtures'
import { ImageRepository } from './ImageRepository'

export default { title: 'Features / Image repositories / Setup' }

export const ECR = () => (
  <PageStory>
    <ImageRepository
      orgId="org-mock-001"
      repository={repositories[0]}
      connection={verifiedECRConnection}
      setup={{
        ...initialECRConnection.setup,
        permissions_cli:
          'aws iam put-role-policy --role-name acme-nuon --policy-name nuon-image-repositories --policy-document file://permissions.json',
      }}
      onVerify={noop}
      onReloadSetup={noop}
    />
  </PageStory>
)
export const GAR = () => (
  <PageStory>
    <ImageRepository
      orgId="org-mock-001"
      repository={repositories[1]}
      connection={verifiedGARConnection}
      setup={{
        ...initialGARConnection.setup,
        permissions_cli:
          'gcloud artifacts repositories add-iam-policy-binding acme-services --project=acme-project --location=us-central1 --role=roles/artifactregistry.reader --member=serviceAccount:nuon-images@acme-project.iam.gserviceaccount.com',
      }}
      onVerify={noop}
      onReloadSetup={noop}
    />
  </PageStory>
)
export const ACR = () => (
  <PageStory>
    <ImageRepository
      orgId="org-mock-001"
      repository={repositories[2]}
      connection={verifiedACRConnection}
      setup={{
        ...initialACRConnection.setup,
        permissions_cli: initialACRConnection.setup.cli.split('# 3.')[1],
      }}
      onVerify={noop}
      onReloadSetup={noop}
    />
  </PageStory>
)
export const Pending = () => (
  <PageStory>
    <ImageRepository
      orgId="org-mock-001"
      repository={repositories[0]}
      connection={initialECRConnection}
      setup={initialECRConnection.setup}
      onVerify={noop}
      onReloadSetup={noop}
    />
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ImageRepository
      orgId="org-mock-001"
      repository={repositories[0]}
      connection={verifiedECRConnection}
      setupError={{
        error: 'Setup generation failed',
        description: 'Try again.',
        user_error: false,
      }}
      onVerify={noop}
      onReloadSetup={noop}
    />
  </PageStory>
)
