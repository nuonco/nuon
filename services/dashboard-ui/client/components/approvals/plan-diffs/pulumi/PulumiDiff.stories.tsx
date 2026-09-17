export default {
  title: 'Approvals/PlanDiffs/PulumiDiff',
}

import {
  s3BucketCreatePlan,
  ecsServiceUpdatePlan,
  databaseReplacePlan,
  mixedInfraChangesPlan,
  largeEksClusterDeployPlan,
  kubernetesAppMigrationPlan,
  azureCosmeticUpdatesPlan,
  rbacArrayNoisePlan,
  longPolicyAndConfigValuesPlan,
  withDiagnosticsPlan,
} from '@/lib/fixtures/plan-diffs/pulumi'
import { PulumiDiff } from './PulumiDiff'

export const S3BucketCreate = () => <PulumiDiff plan={s3BucketCreatePlan} />

export const ECSServiceUpdate = () => <PulumiDiff plan={ecsServiceUpdatePlan} />

export const DatabaseReplace = () => <PulumiDiff plan={databaseReplacePlan} />

export const MixedInfraChanges = () => (
  <PulumiDiff plan={mixedInfraChangesPlan} />
)

export const LargeEKSClusterDeploy = () => (
  <PulumiDiff plan={largeEksClusterDeployPlan} />
)

export const KubernetesAppMigration = () => (
  <PulumiDiff plan={kubernetesAppMigrationPlan} />
)

export const AzureCosmeticUpdates = () => (
  <PulumiDiff plan={azureCosmeticUpdatesPlan} />
)

export const RBACArrayNoise = () => <PulumiDiff plan={rbacArrayNoisePlan} />

export const LongPolicyAndConfigValues = () => (
  <PulumiDiff plan={longPolicyAndConfigValuesPlan} />
)

export const WithDiagnostics = () => <PulumiDiff plan={withDiagnosticsPlan} />
