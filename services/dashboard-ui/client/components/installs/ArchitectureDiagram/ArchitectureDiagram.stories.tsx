import type { TInstall, TInstallComponent, TInstallStack, TAppConfig } from '@/types'
import type { TInstallAppPermissionsConfig } from '@/lib/ctl-api/installs/get-install-app-permissions-config'
import { ArchitectureDiagram } from './ArchitectureDiagram'

export default {
  title: 'Installs/ArchitectureDiagram',
}

const mockInstall: TInstall = {
  id: 'inst-001',
  name: 'production-us-east',
  app_id: 'app-001',
  app_config_id: 'cfg-001',
  sandbox_status: 'active',
  drifted_objects: [],
} as TInstall

const mockComponents: TInstallComponent[] = [
  {
    id: 'ic-001',
    component_id: 'comp-001',
    component: {
      name: 'api-server',
      type: 'helm_chart',
    },
    status_v2: { status: 'healthy', status_human_description: 'Running normally' },
    install_deploys: [
      {
        status_v2: { status: 'succeeded' },
        install_deploy_type: 'deploy',
        created_at: '2026-04-07T10:00:00Z',
      },
    ],
  },
  {
    id: 'ic-002',
    component_id: 'comp-002',
    component: {
      name: 'nginx-ingress',
      type: 'kubernetes_manifest',
    },
    status_v2: { status: 'healthy', status_human_description: 'Running normally' },
    install_deploys: [],
  },
  {
    id: 'ic-003',
    component_id: 'comp-003',
    component: {
      name: 'vpc-network',
      type: 'terraform_module',
    },
    status_v2: { status: 'healthy', status_human_description: 'Provisioned' },
    install_deploys: [
      {
        status_v2: { status: 'succeeded' },
        install_deploy_type: 'deploy',
        created_at: '2026-04-06T08:00:00Z',
      },
    ],
  },
  {
    id: 'ic-004',
    component_id: 'comp-004',
    component: {
      name: 'worker-service',
      type: 'docker_build',
    },
    status_v2: { status: 'deploying', status_human_description: 'Deploy in progress' },
    install_deploys: [],
  },
] as unknown as TInstallComponent[]

const mockStack: TInstallStack = {
  id: 'stack-001',
  versions: [
    {
      composite_status: { status: 'active' },
    },
  ],
} as unknown as TInstallStack

const mockAppConfig: TAppConfig = {
  id: 'cfg-001',
} as TAppConfig

const mockPermissionsConfig: TInstallAppPermissionsConfig = {
  provision_role: {
    id: 'role-001',
    name: 'provision',
    display_name: 'Provision',
    description: 'Role used during resource provisioning',
    type: 'provision',
    policies: [{ name: 'NuonProvisionPolicy', contents: '{}' }],
    permissions_boundary: '',
    created_at: '2026-04-01T00:00:00Z',
    enabled: true,
    arn: 'arn:aws:iam::role/provision',
  },
  deprovision_role: {
    id: 'role-002',
    name: 'deprovision',
    display_name: 'Deprovision',
    description: 'Role used during resource teardown',
    type: 'deprovision',
    policies: [{ name: 'NuonDeprovisionPolicy', contents: '{}' }],
    permissions_boundary: '',
    created_at: '2026-04-01T00:00:00Z',
    enabled: true,
    arn: 'arn:aws:iam::role/deprovision',
  },
  maintenance_role: {
    id: 'role-003',
    name: 'maintenance',
    display_name: 'Maintenance',
    description: 'Role for routine maintenance operations',
    type: 'maintenance',
    policies: [],
    permissions_boundary: '',
    created_at: '2026-04-01T00:00:00Z',
    enabled: false,
    arn: 'arn:aws:iam::role/maintenance',
  },
  break_glass_roles: [],
  custom_roles: [],
}

export const Default = () => (
  <div style={{ width: '100%', height: 600 }}>
    <ArchitectureDiagram
      install={mockInstall}
      components={mockComponents}
      stack={mockStack}
      appConfig={mockAppConfig}
      permissionsConfig={mockPermissionsConfig}
      orgId="org-001"
      isLoading={false}
      isError={false}
    />
  </div>
)

export const Empty = () => (
  <div style={{ width: '100%', height: 600 }}>
    <ArchitectureDiagram
      install={mockInstall}
      components={[]}
      stack={mockStack}
      appConfig={mockAppConfig}
      permissionsConfig={undefined}
      orgId="org-001"
      isLoading={false}
      isError={false}
    />
  </div>
)

export const Loading = () => (
  <div style={{ width: '100%', height: 600 }}>
    <ArchitectureDiagram
      install={mockInstall}
      components={[]}
      orgId="org-001"
      isLoading={true}
      isError={false}
    />
  </div>
)

export const Error = () => (
  <div style={{ width: '100%', height: 600 }}>
    <ArchitectureDiagram
      install={mockInstall}
      components={[]}
      orgId="org-001"
      isLoading={false}
      isError={true}
    />
  </div>
)

export const NoRoles = () => (
  <div style={{ width: '100%', height: 600 }}>
    <ArchitectureDiagram
      install={mockInstall}
      components={mockComponents}
      stack={mockStack}
      appConfig={mockAppConfig}
      permissionsConfig={undefined}
      orgId="org-001"
      isLoading={false}
      isError={false}
    />
  </div>
)
