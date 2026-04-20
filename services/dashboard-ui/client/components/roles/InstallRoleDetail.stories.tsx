import { InstallRoleDetail } from './InstallRoleDetail'

export default {
  title: 'Roles/InstallRoleDetail',
}

const mockInstallRole = {
  id: 'ir-1',
  install_id: 'inst-1',
  app_role_config_id: 'role-1',
  app_role_config: {
    id: 'role-1',
    display_name: 'Provision role',
    description: 'Used to provision infrastructure in customer accounts',
    name: 'nuon-provision-role',
    type: 'provision',
    created_at: '2024-06-15T10:30:00Z',
    enabled: true,
    arn: 'arn:aws:iam::123456789012:role/nuon-provision-role',
    policies: [
      {
        id: 'pol-1',
        managed_policy_name: 'AmazonEKSClusterPolicy',
      },
      {
        id: 'pol-2',
        name: 'custom-s3-access',
        contents: btoa(
          JSON.stringify(
            {
              Version: '2012-10-17',
              Statement: [
                {
                  Effect: 'Allow',
                  Action: ['s3:GetObject'],
                  Resource: 'arn:aws:s3:::my-bucket/*',
                },
              ],
            },
            null,
            2
          )
        ),
      },
    ],
    permissions_boundary: btoa(
      JSON.stringify(
        {
          Version: '2012-10-17',
          Statement: [{ Effect: 'Allow', Action: '*', Resource: '*' }],
        },
        null,
        2
      )
    ),
  },
  enabled: true,
  provisioned: true,
  role_id: 'arn:aws:iam::123456789012:role/nuon-provision-role',
  created_at: '2024-06-15T10:30:00Z',
}

export const Default = () => <InstallRoleDetail installRole={mockInstallRole} />

export const Unprovisioned = () => (
  <InstallRoleDetail
    installRole={{
      ...mockInstallRole,
      provisioned: false,
      role_id: '',
    }}
  />
)

export const NoPolicies = () => (
  <InstallRoleDetail
    installRole={{
      ...mockInstallRole,
      app_role_config: {
        ...mockInstallRole.app_role_config,
        policies: [],
      },
    }}
  />
)
