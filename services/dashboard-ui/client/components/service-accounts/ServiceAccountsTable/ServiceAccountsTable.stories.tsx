export default {
  title: 'ServiceAccounts/ServiceAccountsTable',
}

import { ServiceAccountsTable } from './ServiceAccountsTable'
import type { TAccount } from '@/types'

const mockAccounts: TAccount[] = [
  {
    id: 'acc-1',
    name: 'ci-deploy',
    email: 'svc-ci-deploy@example.com',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    roles: [{ role_type: 'runner' }],
  } as TAccount,
  {
    id: 'acc-2',
    name: 'terraform',
    email: 'svc-terraform@example.com',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
    roles: [{ role_type: 'org_admin' }],
  } as TAccount,
]

const roleTitles = {
  runner: 'Runner',
  org_admin: 'Admin',
  org_read_only: 'Read-only',
}

export const Default = () => (
  <ServiceAccountsTable
    data={mockAccounts}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const WithPagination = () => (
  <ServiceAccountsTable
    data={mockAccounts}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <ServiceAccountsTable
    data={[]}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => (
  <ServiceAccountsTable
    data={[]}
    roleTitles={roleTitles}
    isLoading
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)
