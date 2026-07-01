export default {
  title: 'Installs/Forms/InputConfigFields',
}

import { InputConfigFields } from './InputConfigFields'
import type { TAppInputConfig } from '@/types'

const mockInputConfig: TAppInputConfig = {
  id: 'config-1',
  input_groups: [
    {
      id: 'group-app',
      display_name: 'Application settings',
      description: 'Core application configuration',
      index: 0,
      app_inputs: [
        {
          id: 'app-name',
          name: 'app_name',
          display_name: 'App name',
          description: 'Display name for the application',
          type: 'string',
          required: true,
          default: 'acme-app',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'log-level',
          name: 'log_level',
          display_name: 'Log level',
          description: 'Logging verbosity (debug, info, warn, error)',
          type: 'string',
          required: false,
          default: 'info',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'replica-count',
          name: 'replica_count',
          display_name: 'Replica count',
          description: 'Number of application replicas',
          type: 'number',
          required: false,
          default: '2',
          index: 2,
          source: 'vendor',
        },
        {
          id: 'enable-metrics',
          name: 'enable_metrics',
          display_name: 'Enable metrics',
          description: 'Expose Prometheus metrics',
          type: 'bool',
          required: false,
          default: 'true',
          index: 3,
          source: 'vendor',
        },
        {
          id: 'feature-flags',
          name: 'feature_flags',
          display_name: 'Feature flags',
          description: 'JSON map of feature toggles',
          type: 'json',
          required: false,
          default: '{\n  "new_dashboard": true,\n  "beta_search": false\n}',
          index: 4,
          source: 'vendor',
        },
      ],
    },
    {
      id: 'group-db',
      display_name: 'Database',
      description: 'Database connection settings',
      index: 1,
      app_inputs: [
        {
          id: 'db-host',
          name: 'db_host',
          display_name: 'Database host',
          description: 'The hostname or IP of your database server',
          type: 'string',
          required: true,
          default: 'localhost',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'db-port',
          name: 'db_port',
          display_name: 'Database port',
          description: 'The port number',
          type: 'number',
          required: false,
          default: '5432',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'db-name',
          name: 'db_name',
          display_name: 'Database name',
          description: 'The name of the database',
          type: 'string',
          required: true,
          default: 'acme',
          index: 2,
          source: 'vendor',
        },
        {
          id: 'db-password',
          name: 'db_password',
          display_name: 'Database password',
          description: 'Password for the database user',
          type: 'string',
          required: true,
          sensitive: true,
          index: 3,
          source: 'vendor',
        },
      ],
    },
    {
      id: 'group-net',
      display_name: 'Networking',
      description: 'Ingress and network configuration',
      index: 2,
      app_inputs: [
        {
          id: 'domain',
          name: 'domain',
          display_name: 'Domain',
          description: 'Public domain for the install',
          type: 'string',
          required: true,
          default: 'app.example.com',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'enable-tls',
          name: 'enable_tls',
          display_name: 'Enable TLS',
          description: 'Terminate TLS at the ingress',
          type: 'bool',
          required: false,
          default: 'true',
          index: 1,
          source: 'vendor',
        },
      ],
    },
    {
      id: 'group-overrides',
      name: 'nuon_component_overrides',
      display_name: 'Component overrides',
      description: 'Per-component install-level Helm values and Terraform vars',
      index: 3,
      app_inputs: [
        {
          id: 'input-enabled-whoami',
          name: 'nuon_component_override_v1_enabled_77686f616d69',
          display_name: 'whoami enabled',
          description: 'Whether whoami is deployed on this install',
          type: 'bool',
          required: false,
          default: 'true',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'input-helm-whoami',
          name: 'nuon_component_override_v1_helm_values_77686f616d69',
          display_name: 'whoami Helm values',
          description: 'Raw YAML merged over the component’s app-config values',
          type: 'yaml',
          required: false,
          default:
            'replicaCount: 5\nresources:\n  requests:\n    cpu: "150m"\n    memory: 64Mi\n',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'input-helm-api',
          name: 'nuon_component_override_v1_helm_values_617069',
          display_name: 'api Helm values',
          description: 'Raw YAML merged over the component’s app-config values',
          type: 'yaml',
          required: false,
          default: 'replicaCount: 2\nimage:\n  tag: "2.3.1"\n',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'input-tfvars-vpc',
          name: 'nuon_component_override_v1_tf_vars_767063',
          display_name: 'vpc Terraform vars',
          description: 'Raw .tfvars (HCL or JSON) appended as the final -var-file',
          type: 'hcl',
          required: false,
          default: 'cidr_block = "10.1.0.0/16"\ninstance_count = 3\n',
          index: 2,
          source: 'vendor',
        },
      ],
    },
  ],
} as TAppInputConfig

export const Default = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <InputConfigFields inputConfig={mockInputConfig} />
  </form>
)

export const WithDraftValues = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <InputConfigFields
      inputConfig={mockInputConfig}
      draftValues={{ 'inputs:db_host': 'prod-db.example.com', 'inputs:db_port': '5433' }}
    />
  </form>
)

const codeInputConfig: TAppInputConfig = {
  id: 'config-code',
  input_groups: [
    {
      id: 'group-code',
      display_name: 'Structured inputs',
      description: 'JSON, YAML, and HCL inputs render in a code editor',
      index: 0,
      app_inputs: [
        {
          id: 'code-json',
          name: 'extra_config',
          display_name: 'Extra config (JSON)',
          description: 'Arbitrary JSON configuration',
          type: 'json',
          required: false,
          default: '{\n  "pool": 10,\n  "timeout": "30s"\n}',
          index: 0,
          source: 'vendor',
        },
        {
          id: 'code-yaml',
          name: 'nuon_component_override_v1_helm_values_77686f616d69',
          display_name: 'whoami Helm values (YAML)',
          description: 'Raw YAML merged over the component’s app-config values',
          type: 'yaml',
          required: false,
          default:
            'replicaCount: 5\nresources:\n  requests:\n    cpu: "150m"\n    memory: 64Mi\n',
          index: 1,
          source: 'vendor',
        },
        {
          id: 'code-hcl',
          name: 'nuon_component_override_v1_tf_vars_767063',
          display_name: 'vpc Terraform vars (HCL)',
          description: 'Raw .tfvars (HCL or JSON) appended as the final -var-file',
          type: 'hcl',
          required: false,
          default: 'cidr_block = "10.1.0.0/16"\ninstance_count = 3\n',
          index: 2,
          source: 'vendor',
        },
      ],
    },
  ],
} as TAppInputConfig

export const CodeInputTypes = () => (
  <form className="max-w-2xl p-6 flex flex-col gap-6">
    <InputConfigFields inputConfig={codeInputConfig} />
  </form>
)
