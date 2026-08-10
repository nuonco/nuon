export default {
  title: 'ApiTokens/ApiTokensTable',
}

import type { TStaticToken } from '@/types'
import { ApiTokensTable } from './ApiTokensTable'

const mockTokens: TStaticToken[] = [
  {
    id: 'tok_1',
    name: 'ci-deploy',
    created_at: '2026-06-01T12:00:00Z',
    expires_at: '2027-06-01T12:00:00Z',
  },
  {
    id: 'tok_2',
    name: 'staging-runner',
    created_at: '2026-05-15T09:30:00Z',
    expires_at: '2026-08-15T09:30:00Z',
  },
]

const pagination = { hasNext: false, offset: 0, limit: 20 }

const roleTitles = (roleType: string | undefined) =>
  ({ org_admin: 'Admin', org_support: 'Support', org_read_only: 'Read-only' } as Record<string, string>)[roleType ?? ''] ??
  roleType ??
  '—'

export const Default = () => (
  <ApiTokensTable data={mockTokens} roleTitles={roleTitles} isLoading={false} pagination={pagination} />
)

export const Empty = () => (
  <ApiTokensTable data={[]} roleTitles={roleTitles} isLoading={false} pagination={pagination} />
)

export const Loading = () => (
  <ApiTokensTable data={[]} roleTitles={roleTitles} isLoading={true} pagination={pagination} />
)
