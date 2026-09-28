export default { title: 'Cloud connections/Concepts/Preset flow' }

import { PresetFlow } from './PresetFlow'

export const CloudAndAccount = () => <PresetFlow initialStep={1} />
export const AccessPreset = () => <PresetFlow initialStep={2} />
export const PresetSelected = () => (
  <PresetFlow initialStep={2} initialAccess="preset" />
)
export const CustomSelected = () => (
  <PresetFlow initialStep={2} initialAccess="custom" showTrustPolicy />
)
export const RunInCloudTerraform = () => (
  <PresetFlow
    initialStep={3}
    initialAccess="preset"
    initialFormat="terraform"
  />
)
export const RunInCloudAWSCLI = () => (
  <PresetFlow initialStep={3} initialAccess="preset" initialFormat="cli" />
)
export const RunInCloudCloudFormation = () => (
  <PresetFlow
    initialStep={3}
    initialAccess="preset"
    initialFormat="cloudformation"
  />
)
export const Verify = () => (
  <PresetFlow initialStep={4} initialAccess="preset" />
)
export const Verifying = () => (
  <PresetFlow initialStep={4} initialAccess="preset" verification="verifying" />
)
export const Verified = () => (
  <PresetFlow initialStep={4} initialAccess="preset" verification="verified" />
)
export const FailedAtTrust = () => (
  <PresetFlow
    initialStep={4}
    initialAccess="preset"
    verification="failed-trust"
  />
)
export const FailedAtPermissions = () => (
  <PresetFlow
    initialStep={4}
    initialAccess="preset"
    verification="failed-permissions"
  />
)

CloudAndAccount.meta = { fullBleed: true }
AccessPreset.meta = { fullBleed: true }
PresetSelected.meta = { fullBleed: true }
CustomSelected.meta = { fullBleed: true }
RunInCloudTerraform.meta = { fullBleed: true }
RunInCloudAWSCLI.meta = { fullBleed: true }
RunInCloudCloudFormation.meta = { fullBleed: true }
Verify.meta = { fullBleed: true }
Verifying.meta = { fullBleed: true }
Verified.meta = { fullBleed: true }
FailedAtTrust.meta = { fullBleed: true }
FailedAtPermissions.meta = { fullBleed: true }
