export default {
  title: 'Roles/IAMRoles',
}

import { IAMRoles, IAMRolesSkeleton } from './IAMRoles'
import type { TAppConfig } from '@/types'

const mockAppConfig: TAppConfig = {
  permissions: {
    aws_iam_roles: [
      {
        id: 'role-1',
        display_name: 'Provision role',
        description: 'Used to provision infrastructure for this install.',
        name: 'nuon-provision-role',
        type: 'provision',
        created_at: '2024-01-01T00:00:00Z',
        policies: [
          { id: 'p-1', managed_policy_name: 'AdministratorAccess' },
          { id: 'p-2', name: 'custom-policy', contents: btoa('{"Version":"2012-10-17","Statement":[]}') },
        ],
      },
      {
        id: 'role-2',
        display_name: 'Deprovision role',
        description: 'Used to tear down infrastructure.',
        name: 'nuon-deprovision-role',
        type: 'deprovision',
        created_at: '2024-01-01T00:00:00Z',
        policies: [
          { id: 'p-3', managed_policy_name: 'AdministratorAccess' },
        ],
      },
    ],
  },
} as any

export const Default = () => <IAMRoles appConfig={mockAppConfig} />

export const Loading = () => <IAMRolesSkeleton />

export const NoRoles = () => (
  <IAMRoles
    appConfig={{ permissions: { aws_iam_roles: [] } } as any}
  />
)
