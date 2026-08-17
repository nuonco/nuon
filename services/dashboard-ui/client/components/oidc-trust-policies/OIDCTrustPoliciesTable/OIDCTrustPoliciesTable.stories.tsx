export default {
  title: 'OIDCTrustPolicies/OIDCTrustPoliciesTable',
}

import { OIDCTrustPoliciesTable } from './OIDCTrustPoliciesTable'
import type { TOIDCTrustPolicy } from '@/types'

const mockPolicies: TOIDCTrustPolicy[] = [
  {
    id: 'oidctp-1',
    org_id: 'org-1',
    name: 'GitHub Actions CI',
    issuer_url: 'https://token.actions.githubusercontent.com',
    audience: 'https://api.nuon.co',
    claim_conditions: { sub: 'repo:acme/app:ref:refs/heads/main' },
    role: 'org_read_only',
    token_duration_seconds: 3600,
    enabled: true,
    created_by_id: 'acct-1',
    created_at: '2026-04-20T12:00:00Z',
    updated_at: '2026-04-20T12:00:00Z',
    last_used_at: '2026-07-29T09:00:00Z',
  },
  {
    id: 'oidctp-2',
    org_id: 'org-1',
    name: 'CircleCI deploy',
    issuer_url: 'https://oidc.circleci.com/org/abc-123',
    audience: 'https://api.nuon.co',
    claim_conditions: { sub: 'org/abc-123/project/def-456/*' },
    role: 'org_read_only',
    token_duration_seconds: 3600,
    enabled: false,
    created_by_id: 'acct-1',
    created_at: '2026-04-22T15:30:00Z',
    updated_at: '2026-04-22T15:30:00Z',
  },
]

export const Default = () => (
  <OIDCTrustPoliciesTable data={mockPolicies} isLoading={false} />
)

export const Empty = () => (
  <OIDCTrustPoliciesTable data={[]} isLoading={false} />
)

export const Loading = () => <OIDCTrustPoliciesTable data={[]} isLoading />
