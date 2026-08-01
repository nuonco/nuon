export default {
  title: 'OIDCTrustPolicies/DeleteOIDCTrustPolicy',
}

import { ModalStory } from '@/components/__stories__/helpers'
import { DeleteOIDCTrustPolicyModal } from './DeleteOIDCTrustPolicy'

const noop = () => {}

export const Default = () => (
  <ModalStory>
    <DeleteOIDCTrustPolicyModal
      policyName="GitHub Actions CI"
      isPending={false}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const Pending = () => (
  <ModalStory>
    <DeleteOIDCTrustPolicyModal
      policyName="GitHub Actions CI"
      isPending={true}
      error={null}
      onSubmit={noop}
    />
  </ModalStory>
)

export const WithError = () => (
  <ModalStory>
    <DeleteOIDCTrustPolicyModal
      policyName="GitHub Actions CI"
      isPending={false}
      error={{
        error: 'trust policy not found',
        description: '',
        user_error: true,
        status: 404,
      }}
      onSubmit={noop}
    />
  </ModalStory>
)
