export default { title: 'CloudConnections/CreateCloudConnection' }

import { ModalStory } from '@/components/__stories__/helpers'
import type { TCloudConnection } from '@/types'
import { CreateCloudConnectionModal } from './CreateCloudConnection'

const noop = () => {}

const connection = {
  id: 'cc_01JEXAMPLE',
  name: 'Production AWS',
  platform: 'aws',
  target_id: '123456789012',
  principal: 'arn:aws:iam::123456789012:role/nuon-cloud-connection',
  capabilities: ['stacks', 'images'],
  status: 'pending',
  status_message: '',
  setup: {
    issuer_url: 'https://api.nuon.co',
    subject: 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE',
    audience: 'sts.amazonaws.com',
    trust_policy: {},
    terraform:
      'module "nuon_cloud_connection" {\n  source = "nuonco/ecr-access/aws"\n}',
    cli: 'aws iam create-open-id-connect-provider --url https://api.nuon.co',
    cloudformation: 'NuonConnectionRole:\n  Type: AWS::IAM::Role',
    capabilities: ['stacks', 'images'],
  },
  used_by: { installs: 0, components: 0 },
} as TCloudConnection

const azureConnection = {
  ...connection,
  id: 'cc_01JAZUREEXAMPLE',
  name: 'Production Azure',
  platform: 'azure',
  target_id: '11111111-1111-4111-8111-111111111111',
  tenant_id: '22222222-2222-4222-8222-222222222222',
  principal: '33333333-3333-4333-8333-333333333333',
  setup: {
    ...connection.setup,
    subject: 'org:org_01JEXAMPLE:connection:cc_01JAZUREEXAMPLE',
    audience: 'api://AzureADTokenExchange',
    terraform:
      'module "nuon_cloud_connection" {\n  source = "nuonco/acr-access/azure"\n}',
    cli: 'az ad app create --display-name nuon-cloud-connection',
    cloudformation: '',
    portal_json:
      '{\n  "name": "nuon-cloud-connection",\n  "audiences": ["api://AzureADTokenExchange"]\n}',
    registry: 'acmecontainers',
  },
} as TCloudConnection

const gcpConnection = {
  ...connection,
  id: 'cc_01JGCPEXAMPLE',
  name: 'Production GCP',
  platform: 'gcp',
  target_id: 'acme-production',
  principal: 'nuon-cloud-connection@acme-production.iam.gserviceaccount.com',
  identity_provider:
    'projects/123456789/locations/global/workloadIdentityPools/nuon/providers/connection',
  capabilities: ['images'],
  setup: {
    ...connection.setup,
    subject: 'org:org_01JEXAMPLE:connection:cc_01JGCPEXAMPLE',
    audience:
      'https://iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/nuon/providers/connection',
    terraform:
      'module "nuon_cloud_connection" {\n  source = "nuonco/gar-access/google"\n}',
    cli: 'gcloud iam workload-identity-pools create "nuon" --location global',
    cloudformation: '',
    capabilities: ['images'],
  },
} as TCloudConnection

export const Default = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={null}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const Setup = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={connection}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const AzureCreate = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      defaultPlatform="azure"
      connection={null}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const AzureSetup = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={azureConnection}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const GCPCreate = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      defaultPlatform="gcp"
      connection={null}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)

export const GCPSetup = () => (
  <ModalStory>
    <CreateCloudConnectionModal
      connection={gcpConnection}
      error={null}
      isPending={false}
      isVerifying={false}
      verifyError={null}
      onSubmit={noop}
      onVerify={noop}
      onDone={noop}
    />
  </ModalStory>
)
