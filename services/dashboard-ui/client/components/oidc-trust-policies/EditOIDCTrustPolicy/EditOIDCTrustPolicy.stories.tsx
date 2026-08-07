export default {
  title: 'OIDCTrustPolicies/EditOIDCTrustPolicy',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TOIDCTrustPolicy } from '@/types'
import { EditOIDCTrustPolicyModal } from './EditOIDCTrustPolicy'

const noop = () => {}

const basePolicy: TOIDCTrustPolicy = {
  id: 'oidctp_001',
  org_id: 'org_001',
  name: 'GitHub Actions CI',
  issuer_url: 'https://token.actions.githubusercontent.com',
  audience: 'https://api.nuon.co',
  claim_conditions: { sub: 'repo:acme/app:ref:refs/heads/main' },
  role: 'org_support',
  token_duration_seconds: 3600,
  enabled: true,
  created_at: '2026-04-01T00:00:00Z',
  updated_at: '2026-04-01T00:00:00Z',
}

const roleOptions = [
  { value: 'org_read_only', label: 'org_read_only' },
  { value: 'org_support', label: 'org_support' },
  { value: 'org_admin', label: 'org_admin' },
]

export const Default = () => (
  <ModalStory>
    <EditOIDCTrustPolicyModal
      policy={basePolicy}
      isPending={false}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Disabled = () => (
  <ModalStory>
    <EditOIDCTrustPolicyModal
      policy={{ ...basePolicy, enabled: false }}
      isPending={false}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <EditOIDCTrustPolicyModal
      policy={basePolicy}
      isPending={true}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <EditOIDCTrustPolicyModal
      policy={basePolicy}
      isPending={false}
      error={{
        error: 'claim_conditions must include a "sub" condition',
        description: '',
        user_error: true,
        status: 400,
      }}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)
