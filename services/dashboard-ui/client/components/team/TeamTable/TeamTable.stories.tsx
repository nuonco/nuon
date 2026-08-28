export default {
  title: 'Team/TeamTable',
}

import { TeamTable } from './TeamTable'
import type { TAccount } from '@/types'

const mockAccounts: TAccount[] = [
  {
    id: 'acc-1',
    email: 'alice@example.com',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  } as TAccount,
  {
    id: 'acc-2',
    email: 'bob@example.com',
    created_at: '2024-01-02T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z',
  } as TAccount,
  {
    id: 'acc-3',
    email: 'charlie@example.com',
    created_at: '2024-01-03T00:00:00Z',
    updated_at: '2024-01-03T00:00:00Z',
  } as TAccount,
]

const roleTitles = (roleType: string | undefined) =>
  (
    ({
      org_admin: 'Admin',
      org_support: 'Support',
      org_read_only: 'Read-only',
    }) as Record<string, string>
  )[roleType ?? ''] ??
  roleType ??
  '—'

export const Default = () => (
  <TeamTable
    data={mockAccounts}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const WithPagination = () => (
  <TeamTable
    data={mockAccounts}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <TeamTable
    data={[]}
    roleTitles={roleTitles}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => (
  <TeamTable
    data={[]}
    roleTitles={roleTitles}
    isLoading
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)
