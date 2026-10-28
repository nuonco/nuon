export default {
  title: 'Features / API tokens / API tokens table',
}

import type { TToken } from '@/types'
import { ApiTokensTable } from './ApiTokensTable'

const mockTokens: TToken[] = [
  {
    id: 'tok_1',
    name: 'ci-deploy',
    account_id: 'acct_sa_1',
    account_name: 'ci-deploy-bot',
    account_email: 'ci-deploy-bot@acme.svc.nuon.co',
    account_role: 'Read-only',
    created_at: '2026-06-01T12:00:00Z',
    expires_at: '2027-06-01T12:00:00Z',
  },
  {
    id: 'tok_2',
    name: 'release-automation',
    account_id: 'acct_sa_2',
    account_name: 'release-automation',
    account_email: 'release-automation@acme.svc.nuon.co',
    account_role: 'Admin',
    created_at: '2026-05-15T09:30:00Z',
    expires_at: '2026-08-15T09:30:00Z',
  },
]

const pagination = { hasNext: false, offset: 0, limit: 20 }

export const Default = () => (
  <ApiTokensTable data={mockTokens} isLoading={false} pagination={pagination} />
)

export const Empty = () => (
  <ApiTokensTable data={[]} isLoading={false} pagination={pagination} />
)

export const Loading = () => (
  <ApiTokensTable data={[]} isLoading={true} pagination={pagination} />
)
