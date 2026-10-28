export default {
  title: 'Features / API tokens / Legacy personal tokens',
}

import type { TStaticToken } from '@/types'
import { LegacyPersonalTokens as LegacyPersonalTokensComponent } from './LegacyPersonalTokens'

const mockTokens: TStaticToken[] = [
  {
    id: 'tok_1',
    name: 'my-personal-token',
    role: 'org_admin',
    account_id: 'acct_user_1',
    created_by_id: 'acct_user_1',
    created_at: '2026-05-15T09:30:00Z',
    expires_at: '2026-08-15T09:30:00Z',
  },
]

const noop = () => {}

export const Default = () => (
  <LegacyPersonalTokensComponent data={mockTokens} isLoading={false} onRevoke={noop} />
)

export const Loading = () => (
  <LegacyPersonalTokensComponent data={[]} isLoading={true} onRevoke={noop} />
)

export const Hidden = () => (
  <LegacyPersonalTokensComponent data={[]} isLoading={false} onRevoke={noop} />
)
