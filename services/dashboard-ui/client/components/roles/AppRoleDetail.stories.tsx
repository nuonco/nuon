import { AppRoleDetail } from './AppRoleDetail'

export default {
  title: 'Roles/AppRoleDetail',
}

const mockRole = {
  id: 'role-1',
  display_name: 'Provision role',
  description: 'Used to provision infrastructure in customer accounts',
  name: 'nuon-provision-role',
  type: 'provision',
  created_at: '2024-06-15T10:30:00Z',
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
                Action: ['s3:GetObject', 's3:PutObject'],
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
        Statement: [
          {
            Effect: 'Allow',
            Action: '*',
            Resource: '*',
          },
        ],
      },
      null,
      2
    )
  ),
}

export const Default = () => <AppRoleDetail role={mockRole} />

export const NoPolicies = () => (
  <AppRoleDetail role={{ ...mockRole, policies: [] }} />
)

export const NoBoundary = () => (
  <AppRoleDetail role={{ ...mockRole, permissions_boundary: undefined }} />
)
