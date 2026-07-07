export default {
  title: 'Installs/UpdateInstallForm',
}

import { UpdateInstallForm } from './UpdateInstallForm'

const mockInstall = {
  id: 'install-1',
  name: 'my-install',
} as any

const mockInputConfig = {
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
          description: 'Logging verbosity',
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
          description: 'Hostname or IP of the database',
          type: 'string',
          required: true,
          default: 'localhost',
          index: 0,
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
          index: 1,
          source: 'vendor',
        },
      ],
    },
    {
      id: 'group-overrides',
      display_name: 'Component overrides',
      description: 'Per-component Helm values and Terraform vars',
      index: 2,
      app_inputs: [
        {
          id: 'input-helm',
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
          id: 'input-tfvars',
          name: 'nuon_component_override_v1_tf_vars_767063',
          display_name: 'vpc Terraform vars',
          description: 'Raw .tfvars (HCL or JSON) appended as the final -var-file',
          type: 'hcl',
          required: false,
          default: 'cidr_block = "10.1.0.0/16"\ninstance_count = 3\n',
          index: 1,
          source: 'vendor',
        },
      ],
    },
  ],
} as any

export const Default = () => (
  <div className="max-w-2xl p-4">
    <UpdateInstallForm
      install={mockInstall}
      platform="aws"
      inputConfig={mockInputConfig}
      onCancel={() => {}}
    />
  </div>
)

export const WithoutInputConfig = () => (
  <div className="max-w-2xl p-4">
    <UpdateInstallForm
      install={mockInstall}
      platform="aws"
      onCancel={() => {}}
    />
  </div>
)
