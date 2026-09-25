export default { title: 'Cloud connections/Concepts/Guided wizard' }

import { GuidedWizard } from './GuidedWizard'

export const CloudAndAccount = () => <GuidedWizard initialStep={1} />
export const Capabilities = () => <GuidedWizard initialStep={2} />
export const CustomizeAccess = () => <GuidedWizard initialStep={3} />
export const RunInAWS = () => <GuidedWizard initialStep={4} />
export const Verifying = () => (
  <GuidedWizard initialStep={5} verifyState="verifying" />
)
export const Verified = () => (
  <GuidedWizard initialStep={5} verifyState="verified" />
)
export const VerificationFailed = () => (
  <GuidedWizard initialStep={5} verifyState="failed" />
)

CloudAndAccount.meta = { fullBleed: true }
Capabilities.meta = { fullBleed: true }
CustomizeAccess.meta = { fullBleed: true }
RunInAWS.meta = { fullBleed: true }
Verifying.meta = { fullBleed: true }
Verified.meta = { fullBleed: true }
VerificationFailed.meta = { fullBleed: true }
