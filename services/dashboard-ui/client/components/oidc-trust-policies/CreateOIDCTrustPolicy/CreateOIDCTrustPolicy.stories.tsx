export default {
  title: 'OIDCTrustPolicies/CreateOIDCTrustPolicy',
}

import { ModalStory } from '@/components/__stories__/helpers'
import type { TVCSConnectionRepo } from '@/types'
import { CreateOIDCTrustPolicyModal } from './CreateOIDCTrustPolicy'

const noop = () => {}

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

const baseProps = {
  isPending: false,
  error: null,
  onSubmit: noop,
  repos,
  hasVCSConnections: true,
  vcsConnectionsHref: '/org-123/connections/vcs',
  githubAudience: 'https://api.nuon.co',
}

export const Default = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal {...baseProps} />
  </ModalStory>
)

export const RepoLocked = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal
      {...baseProps}
      initialRepoFullName="acme/api"
      initialRepoDefaultBranch="main"
      lockPreset
    />
  </ModalStory>
)

export const NoVCSConnections = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal
      {...baseProps}
      repos={[]}
      hasVCSConnections={false}
    />
  </ModalStory>
)

export const LoadingRepos = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal {...baseProps} repos={[]} isLoadingRepos />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal {...baseProps} isPending />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <CreateOIDCTrustPolicyModal
      {...baseProps}
      error={{
        error: 'claim_conditions must include a "sub" condition',
        description: '',
        user_error: true,
        status: 400,
      }}
    />
  </ModalStory>
)
