import { ConnectionForm } from './ConnectionForm'
import { ConnectionWizard } from './ConnectionWizard'
import { PageStory, noop } from '../__stories__/fixtures'

export default { title: 'Features / Cloud connections / Connection form' }

export const Empty = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm onCreate={noop} />
    </ConnectionWizard>
  </PageStory>
)
export const ECR = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        values={{
          name: 'acme-ecr-images',
          target_id: '123456789012',
          ecr_region: 'us-west-2',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const GAR = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        values={{
          name: 'acme-gar-images',
          platform: 'gcp',
          target_id: 'acme-project',
          gcp_project_number: '123456789012',
          gar_location: 'us-central1',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const ACR = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        values={{
          name: 'acme-acr-images',
          platform: 'azure',
          target_id: '11111111-1111-4111-8111-111111111111',
          azure_tenant_id: '22222222-2222-4222-8222-222222222222',
          principal: '33333333-3333-4333-8333-333333333333',
          azure_resource_group: 'acme-images',
          azure_registry_name: 'acmeimages',
          azure_registry_login_server: 'acmeimages-abc123.azurecr.io',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const Stacks = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        values={{
          name: 'acme-stacks',
          preset: 'stacks',
          target_id: '123456789012',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const Custom = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        values={{
          name: 'acme-custom',
          preset: 'custom',
          target_id: '123456789012',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const Creating = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        isPending
        values={{
          name: 'acme-images',
          target_id: '123456789012',
          ecr_region: 'us-west-2',
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
export const Failed = () => (
  <PageStory>
    <ConnectionWizard>
      <ConnectionForm
        onCreate={noop}
        error={{
          error: 'Connection creation failed',
          description: 'This identity already has a cloud connection.',
          user_error: true,
        }}
      />
    </ConnectionWizard>
  </PageStory>
)
