export default {
  title: 'OIDCTrustPolicies/OIDCTrustPolicyForm',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TOIDCTrustPolicy, TVCSConnectionRepo } from '@/types'
import { OIDCTrustPolicyFormModal } from './OIDCTrustPolicyForm'

const noop = () => {}

const roleOptions = [
  { value: 'org_read_only', label: 'org_read_only' },
  { value: 'org_support', label: 'org_support' },
  { value: 'org_admin', label: 'org_admin' },
]

const repos = [
  {
    id: 1,
    name: 'api',
    full_name: 'acme/api',
    private: true,
    fork: false,
    html_url: 'https://github.com/acme/api',
    default_branch: 'main',
    updated_at: '2026-07-01T00:00:00Z',
  },
  {
    id: 2,
    name: 'infra',
    full_name: 'acme/infra',
    private: true,
    fork: false,
    html_url: 'https://github.com/acme/infra',
    default_branch: 'trunk',
    updated_at: '2026-07-02T00:00:00Z',
  },
] satisfies TVCSConnectionRepo[]

const createProps = {
  mode: 'create' as const,
  isPending: false,
  error: null,
  onSubmit: noop,
  roleOptions,
  repos,
  hasVCSConnections: true,
  vcsConnectionsHref: '/org-123/connections/vcs',
  githubAudience: 'https://api.nuon.co',
}

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

export const Create = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal {...createProps} />
  </ModalStory>
)

export const CreateRepoLocked = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      {...createProps}
      initialRepoFullName="acme/api"
      initialRepoDefaultBranch="main"
      lockPreset
    />
  </ModalStory>
)

export const CreateNoVCSConnections = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      {...createProps}
      repos={[]}
      hasVCSConnections={false}
    />
  </ModalStory>
)

export const CreateLoadingRepos = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal {...createProps} repos={[]} isLoadingRepos />
  </ModalStory>
)

export const CreatePending = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal {...createProps} isPending />
  </ModalStory>
)

export const CreateWithError = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      {...createProps}
      error={{
        error: 'claim_conditions must include a "sub" condition',
        description: '',
        user_error: true,
        status: 400,
      }}
    />
  </ModalStory>
)

export const Edit = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      mode="edit"
      policy={basePolicy}
      isPending={false}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditDisabled = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      mode="edit"
      policy={{ ...basePolicy, enabled: false }}
      isPending={false}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditPending = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      mode="edit"
      policy={basePolicy}
      isPending={true}
      error={null}
      roleOptions={roleOptions}
      onSubmit={noop}
    />
  </ModalStory>
)

export const EditWithError = () => (
  <ModalStory>
    <OIDCTrustPolicyFormModal
      mode="edit"
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
