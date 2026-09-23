export default {
  title: 'Installs/Configuration/InstallConfigInputs',
}

import { Button } from '@/components/common/Button'
import type { TAppInput } from '@/types'
import { InstallConfigInputs } from './InstallConfigInputs'

const input = (overrides: Partial<TAppInput>): TAppInput =>
  ({
    id: 'input-1',
    name: 'cluster_size',
    display_name: 'Cluster size',
    default: '3',
    ...overrides,
  }) as TAppInput

const cloudGroup = {
  id: 'group-cloud',
  name: 'cloud',
  display_name: 'Cloud',
  description: 'Where this install runs.',
  app_inputs: [
    input({
      id: 'in-region',
      name: 'region',
      display_name: 'Region',
      default: 'us-west-2',
    }),
    input({
      id: 'in-account',
      name: 'account_id',
      display_name: 'AWS account',
      default: undefined,
    }),
  ],
}

const platformGroup = {
  id: 'group-platform',
  name: 'platform',
  display_name: 'Platform',
  app_inputs: [
    input({ id: 'in-size', name: 'cluster_size' }),
    input({
      id: 'in-domain',
      name: 'domain',
      display_name: 'Domain',
      default: 'example.com',
    }),
    input({ id: 'in-raw', name: 'feature_flags', display_name: undefined }),
  ],
}

const secretsGroup = {
  id: 'group-secrets',
  name: 'secrets',
  display_name: 'Secrets',
  description: 'Values are redacted after they are written.',
  app_inputs: [
    input({
      id: 'in-token',
      name: 'api_token',
      display_name: 'API token',
      default: undefined,
      sensitive: true,
    }),
  ],
}

const editInputs = <Button variant="secondary">Edit inputs</Button>

const managedByConfigAction = (
  <Button
    variant="secondary"
    disabled
    tooltipProps={{
      tipContent: 'Managed by config. Disable config sync to edit.',
    }}
  >
    Edit inputs
  </Button>
)

export const Default = () => (
  <InstallConfigInputs
    action={editInputs}
    groups={[cloudGroup, platformGroup, secretsGroup]}
    values={{
      region: 'us-west-2',
      account_id: '000000000000',
      cluster_size: '5',
      domain: 'payments.example.com',
      feature_flags: '',
      api_token: '***',
    }}
  />
)

export const MissingAndEmptyValues = () => (
  <InstallConfigInputs
    action={editInputs}
    groups={[platformGroup]}
    values={{ cluster_size: '', domain: 'payments.example.com' }}
  />
)

export const LongValues = () => (
  <InstallConfigInputs
    action={editInputs}
    groups={[
      {
        ...platformGroup,
        app_inputs: [
          input({
            id: 'in-conn',
            name: 'connection_string',
            display_name: 'Connection string',
            default: undefined,
          }),
          input({
            id: 'in-cert',
            name: 'ca_bundle',
            display_name: 'CA bundle',
            default: undefined,
          }),
        ],
      },
    ]}
    values={{
      connection_string: `postgres://service-account@${'db'.repeat(40)}.example.com:5432/payments?sslmode=verify-full`,
      ca_bundle: 'a'.repeat(600),
    }}
  />
)

export const ManagedByConfig = () => (
  <InstallConfigInputs
    action={managedByConfigAction}
    groups={[cloudGroup, platformGroup]}
    values={{ region: 'us-west-2', cluster_size: '5' }}
  />
)

export const Loading = () => <InstallConfigInputs groups={[]} isLoading />

export const Empty = () => <InstallConfigInputs groups={[]} />
