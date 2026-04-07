export default {
  title: 'Team/TeamTable',
}

import { TeamTable, TeamTableSkeleton } from './TeamTable'
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

export const Default = () => (
  <TeamTable
    data={mockAccounts}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const WithPagination = () => (
  <TeamTable
    data={mockAccounts}
    isLoading={false}
    pagination={{ hasNext: true, offset: 0, limit: 20 }}
  />
)

export const Empty = () => (
  <TeamTable
    data={[]}
    isLoading={false}
    pagination={{ hasNext: false, offset: 0, limit: 20 }}
  />
)

export const Loading = () => <TeamTableSkeleton />
