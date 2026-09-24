import {
  azureNoOpWithCosmeticDriftPlan,
  driftDetectedPlan,
  driftWithChangesAndOutputsPlan,
  eksClusterCreatePlan,
  iamRoleWithNestedPolicyPlan,
  kubectlManifestDeploymentPlan,
  mixedWithNoOpPlan,
  noOpAndReadResourcesPlan,
  rbacArrayNoisePlan,
  rdsReplacePlan,
  replaceResourcesPlan,
  securityGroupUpdatePlan,
  withPlan,
} from '@/lib/fixtures/plan-diffs/terraform'
import { TerraformDiff } from './TerraformDiff'

export default {
  title: 'Approvals/PlanDiffs/TerraformDiff',
}

export const NoPlan = () => <TerraformDiff plan={undefined} />

export const WithPlan = () => <TerraformDiff plan={withPlan} />

export const IAMRoleWithNestedPolicy = () => (
  <TerraformDiff plan={iamRoleWithNestedPolicyPlan} />
)

export const EKSClusterCreate = () => (
  <TerraformDiff plan={eksClusterCreatePlan} />
)

export const SecurityGroupUpdate = () => (
  <TerraformDiff plan={securityGroupUpdatePlan} />
)

export const RDSReplace = () => <TerraformDiff plan={rdsReplacePlan} />

export const NoOpAndReadResources = () => (
  <TerraformDiff plan={noOpAndReadResourcesPlan} />
)

export const ReplaceResources = () => (
  <TerraformDiff plan={replaceResourcesPlan} />
)

export const MixedWithNoOp = () => <TerraformDiff plan={mixedWithNoOpPlan} />

export const RBACArrayNoise = () => <TerraformDiff plan={rbacArrayNoisePlan} />

export const AzureNoOpWithCosmeticDrift = () => (
  <TerraformDiff plan={azureNoOpWithCosmeticDriftPlan} />
)

export const DriftDetected = () => <TerraformDiff plan={driftDetectedPlan} />

export const KubectlManifestDeployment = () => (
  <TerraformDiff plan={kubectlManifestDeploymentPlan} />
)

export const DriftWithChangesAndOutputs = () => (
  <TerraformDiff plan={driftWithChangesAndOutputsPlan} />
)
