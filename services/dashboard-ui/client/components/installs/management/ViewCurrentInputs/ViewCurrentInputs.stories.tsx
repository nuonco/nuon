export default {
  title: 'Installs/ViewCurrentInputs',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { ViewCurrentInputsModal } from './ViewCurrentInputs'

const mockInputGroups = [
  {
    id: 'group-app',
    display_name: 'Application settings',
    description: 'Core application configuration',
    app_inputs: [
      {
        name: 'app_name',
        display_name: 'App name',
        description: 'Display name for the application',
        required: true,
        sensitive: false,
        source: 'vendor',
        default: 'acme-app',
      },
      {
        name: 'log_level',
        display_name: 'Log level',
        description: 'Logging verbosity',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: 'info',
      },
      {
        name: 'replica_count',
        display_name: 'Replica count',
        description: 'Number of application replicas',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: '2',
      },
      {
        name: 'enable_metrics',
        display_name: 'Enable metrics',
        description: 'Expose Prometheus metrics',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: 'true',
      },
      {
        name: 'feature_flags',
        display_name: 'Feature flags',
        description: 'JSON map of feature toggles',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: '{}',
      },
    ],
  },
  {
    id: 'group-db',
    display_name: 'Database',
    description: 'Database connection settings',
    app_inputs: [
      {
        name: 'db_host',
        display_name: 'Database host',
        description: 'Hostname or IP of the database',
        required: true,
        sensitive: false,
        source: 'vendor',
        default: 'localhost',
      },
      {
        name: 'db_port',
        display_name: 'Database port',
        description: 'Port the database listens on',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: '5432',
      },
      {
        name: 'db_name',
        display_name: 'Database name',
        description: 'Name of the database',
        required: true,
        sensitive: false,
        source: 'vendor',
        default: 'acme',
      },
      {
        name: 'db_username',
        display_name: 'Database user',
        description: 'Database username',
        required: true,
        sensitive: false,
        source: 'vendor',
        default: 'acme',
      },
      {
        name: 'db_password',
        display_name: 'Database password',
        description: 'Database password',
        required: true,
        sensitive: true,
        source: 'vendor',
      },
    ],
  },
  {
    id: 'group-net',
    display_name: 'Networking',
    description: 'Ingress and network configuration',
    app_inputs: [
      {
        name: 'domain',
        display_name: 'Domain',
        description: 'Public domain for the install',
        required: true,
        sensitive: false,
        source: 'vendor',
        default: 'app.example.com',
      },
      {
        name: 'enable_tls',
        display_name: 'Enable TLS',
        description: 'Terminate TLS at the ingress',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: 'true',
      },
      {
        name: 'allowed_cidrs',
        display_name: 'Allowed CIDRs',
        description: 'Comma-separated list of allowed CIDRs',
        required: false,
        sensitive: false,
        source: 'vendor',
        default: '0.0.0.0/0',
      },
    ],
  },
  {
    id: 'group-customer',
    display_name: 'Customer configuration',
    description: 'Values provided by the customer',
    app_inputs: [
      {
        name: 'customer_account_id',
        display_name: 'AWS account ID',
        description: 'Customer AWS account ID',
        required: true,
        sensitive: false,
        source: 'customer',
      },
      {
        name: 'customer_region',
        display_name: 'AWS region',
        description: 'Region to deploy into',
        required: true,
        sensitive: false,
        source: 'customer',
      },
    ],
  },
]

const mockRedacted = {
  app_name: 'acme-prod',
  log_level: 'debug',
  replica_count: '3',
  enable_metrics: 'true',
  feature_flags: '{"new_dashboard":true,"beta_search":false}',
  db_host: 'prod-db.acme.internal',
  db_port: '5432',
  db_name: 'acme_production',
  db_username: 'acme_app',
  db_password: '****',
  domain: 'app.acme.com',
  enable_tls: 'true',
  allowed_cidrs: '10.0.0.0/8,192.168.0.0/16',
  customer_account_id: '123456789012',
  customer_region: 'us-east-1',
}

const mockOverrideGroups = [
  {
    id: 'group-overrides',
    display_name: 'Component overrides',
    description: 'Per-component Helm values and Terraform vars',
    app_inputs: [
      {
        name: 'nuon_component_override_v1_helm_values_77686f616d69',
        display_name: 'whoami Helm values',
        description: 'YAML merged over the component’s app-config values',
        required: false,
        sensitive: false,
        source: 'vendor',
      },
      {
        name: 'nuon_component_override_v1_helm_values_617069',
        display_name: 'api Helm values',
        description: 'YAML merged over the component’s app-config values',
        required: false,
        sensitive: false,
        source: 'vendor',
      },
      {
        name: 'nuon_component_override_v1_tf_vars_767063',
        display_name: 'vpc Terraform vars',
        description: '.tfvars appended as the final -var-file',
        required: false,
        sensitive: false,
        source: 'vendor',
      },
    ],
  },
]

const mockOverrideRedacted = {
  nuon_component_override_v1_helm_values_77686f616d69:
    'replicaCount: 5\nresources:\n  requests:\n    cpu: "150m"\n    memory: 64Mi\n',
  nuon_component_override_v1_helm_values_617069:
    'replicaCount: 2\nimage:\n  tag: "2.3.1"\n',
  nuon_component_override_v1_tf_vars_767063:
    'cidr_block = "10.1.0.0/16"\ninstance_count = 3\n',
}

const mockLongOverrideRedacted = {
  nuon_component_override_v1_helm_values_77686f616d69: `replicaCount: 3
image:
  repository: ghcr.io/acme/whoami
  tag: "1.24.0"
  pullPolicy: IfNotPresent
resources:
  requests:
    cpu: "150m"
    memory: 64Mi
  limits:
    cpu: "500m"
    memory: 256Mi
service:
  type: ClusterIP
  port: 80
ingress:
  enabled: true
  className: nginx
  hosts:
    - host: whoami.example.com
      paths:
        - path: /
          pathType: Prefix
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 75
nodeSelector:
  kubernetes.io/os: linux
`,
  nuon_component_override_v1_helm_values_617069: `replicaCount: 2
image:
  repository: ghcr.io/acme/api
  tag: "2.3.1"
env:
  - name: LOG_LEVEL
    value: debug
  - name: FEATURE_X
    value: "true"
`,
  nuon_component_override_v1_tf_vars_767063: `region             = "us-east-1"
cidr_block         = "10.1.0.0/16"
instance_count     = 3
instance_type      = "t3.large"
enable_nat_gateway = true
availability_zones = ["us-east-1a", "us-east-1b", "us-east-1c"]
private_subnets    = ["10.1.1.0/24", "10.1.2.0/24", "10.1.3.0/24"]
public_subnets     = ["10.1.101.0/24", "10.1.102.0/24", "10.1.103.0/24"]
tags = {
  environment = "production"
  team        = "platform"
  managed_by  = "nuon"
}
`,
}

export const WithGroups = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={mockRedacted}
      inputGroups={mockInputGroups}
    />
  </ModalStory>
)

export const WithOverrides = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={{ ...mockRedacted, ...mockOverrideRedacted }}
      inputGroups={[...mockInputGroups, ...mockOverrideGroups]}
    />
  </ModalStory>
)

export const Loading = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={true}
      redactedValues={{}}
      inputGroups={[]}
    />
  </ModalStory>
)

export const Empty = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={{}}
      inputGroups={[]}
    />
  </ModalStory>
)

export const FlatInputs = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={{ key1: 'value1', key2: 'value2' }}
      inputGroups={[]}
    />
  </ModalStory>
)

export const ComponentOverrides = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={mockOverrideRedacted}
      inputGroups={mockOverrideGroups}
    />
  </ModalStory>
)

export const ComponentOverridesLongValues = () => (
  <ModalStory>
    <ViewCurrentInputsModal
      isLoading={false}
      redactedValues={mockLongOverrideRedacted}
      inputGroups={mockOverrideGroups}
    />
  </ModalStory>
)
