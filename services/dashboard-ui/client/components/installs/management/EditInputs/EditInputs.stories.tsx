export default {
  title: 'Installs/EditInputs',
}

import { useRef } from 'react'
import { ModalStory } from '@/components/__stories__/helpers'
import { EditInputsFormModal } from './EditInputs'

const noop = () => {}

const mockInstall = {
  id: 'inst-123',
  name: 'prod-acme',
  app_id: 'app-123',
  app_config_id: 'cfg-123',
} as any

const mockConfig = {
  id: 'cfg-123',
  input: {
    input_groups: [
      {
        id: 'group-app',
        display_name: 'Application settings',
        description: 'Core application configuration',
        index: 0,
      },
      {
        id: 'group-db',
        display_name: 'Database',
        description: 'Database connection settings',
        index: 1,
      },
      {
        id: 'group-net',
        display_name: 'Networking',
        description: 'Ingress and network configuration',
        index: 2,
      },
      {
        id: 'group-overrides',
        display_name: 'Component overrides',
        description: 'Per-component Helm values and Terraform vars',
        index: 3,
      },
    ],
    inputs: [
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
        group_id: 'group-app',
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
        group_id: 'group-app',
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
        group_id: 'group-app',
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
        group_id: 'group-app',
      },
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
        group_id: 'group-db',
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
        group_id: 'group-db',
      },
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
        group_id: 'group-net',
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
        group_id: 'group-net',
      },
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
        group_id: 'group-overrides',
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
        group_id: 'group-overrides',
      },
    ],
  },
} as any

export const Loading = () => {
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  return (
    <ModalStory>
      <EditInputsFormModal
        install={mockInstall}
        config={undefined}
        isLoading={true}
        error={null}
        isSubmitting={false}
        actionError={null}
        onFormSubmit={noop}
        onClose={noop}
        formRef={formRef}
        clearDraftRef={clearDraftRef}
        selectedRole=""
        onRoleChange={noop}
        deployDependents={true}
        onDeployDependentsChange={noop}
        onMutate={noop as any}
      />
    </ModalStory>
  )
}

export const WithInputs = () => {
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  return (
    <ModalStory>
      <EditInputsFormModal
        install={mockInstall}
        config={mockConfig}
        isLoading={false}
        error={null}
        isSubmitting={false}
        actionError={null}
        onFormSubmit={noop}
        onClose={noop}
        formRef={formRef}
        clearDraftRef={clearDraftRef}
        selectedRole=""
        onRoleChange={noop}
        deployDependents={true}
        onDeployDependentsChange={noop}
        onMutate={noop as any}
      />
    </ModalStory>
  )
}

export const WithError = () => {
  const formRef = useRef<HTMLFormElement>(null)
  const clearDraftRef = useRef<(() => void) | null>(null)

  return (
    <ModalStory>
      <EditInputsFormModal
        install={mockInstall}
        config={undefined}
        isLoading={false}
        error={{ error: 'Unable to load app configuration' }}
        isSubmitting={false}
        actionError={null}
        onFormSubmit={noop}
        onClose={noop}
        formRef={formRef}
        clearDraftRef={clearDraftRef}
        selectedRole=""
        onRoleChange={noop}
        deployDependents={true}
        onDeployDependentsChange={noop}
        onMutate={noop as any}
      />
    </ModalStory>
  )
}
