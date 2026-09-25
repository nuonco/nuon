export default { title: 'Cloud connections/Concepts/Consent grant' }

import { ConsentGrant } from './ConsentGrant'

export const Default = () => <ConsentGrant />
export const ImagesOnly = () => (
  <ConsentGrant
    initialStacks={false}
    initialRepos={['acme/api', 'acme/worker']}
  />
)
export const CustomPolicyOpen = () => (
  <ConsentGrant customOpen initialRepos={['acme/api', 'acme/worker']} />
)
export const Verified = () => <ConsentGrant result="verified" />
export const VerificationFailed = () => <ConsentGrant result="failed" />

Default.meta = { fullBleed: true }
ImagesOnly.meta = { fullBleed: true }
CustomPolicyOpen.meta = { fullBleed: true }
Verified.meta = { fullBleed: true }
VerificationFailed.meta = { fullBleed: true }
