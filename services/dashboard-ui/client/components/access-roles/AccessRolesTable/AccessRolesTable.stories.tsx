import type { TRoleInfo } from '@/types'
import { AccessRolesTable } from './AccessRolesTable'

export default {
  title: 'Access roles/AccessRolesTable',
}

const ORG_ID = 'orgrok933tcyzji01s7us3aeo3'

const mockRoles: TRoleInfo[] = [
  {
    id: 'rolq7fplr1up5atx5zpxotbabm',
    role_type: 'org_admin',
    title: 'Admin',
    description: 'Full access to everything in the org.',
    applies_to: ['team', 'service_account', 'api_token'],
    managed: true,
  },
  {
    id: 'rol4plkdhwau58atwfd92vlc8q',
    role_type: 'custom',
    title: 'Release manager',
    description: 'Reads everything, deploys to production only.',
    applies_to: ['team', 'api_token'],
    managed: false,
    policies: [
      {
        scoped_permissions: [
          { resource_type: 'org', resource_id: ORG_ID, permissions: ['read'] },
          { resource_type: 'app', resource_id: '*', permissions: ['read'] },
          { resource_type: 'install', resource_id: '*', permissions: ['read'] },
          {
            resource_type: 'install',
            resource_id: 'inl4plkdhwau58atwfd92vlc8q',
            permissions: ['create', 'read', 'update', 'delete'],
          },
        ],
      },
    ],
  },
  {
    id: 'rolz7fplr1up5atx5zpxotbabm',
    role_type: 'custom',
    title: 'Unassignable',
    description: 'Has permissions but no assignment surfaces.',
    applies_to: [],
    managed: false,
    policies: [
      {
        scoped_permissions: [
          { resource_type: 'app', resource_id: '*', permissions: ['read'] },
        ],
      },
    ],
  },
]

const nameFor = (id: string) =>
  ({ inl4plkdhwau58atwfd92vlc8q: 'acme-prod' })[id]

export const Default = () => (
  <AccessRolesTable data={mockRoles} isLoading={false} nameFor={nameFor} />
)

export const Loading = () => <AccessRolesTable data={[]} isLoading />

export const Empty = () => <AccessRolesTable data={[]} isLoading={false} />
