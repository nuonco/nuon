import {
  azureCosmeticUpdatesPlan,
  databaseReplacePlan,
  ecsServiceUpdatePlan,
  kubernetesAppMigrationPlan,
  largeEksClusterDeployPlan,
  longPolicyAndConfigValuesPlan,
  mixedInfraChangesPlan,
  rbacArrayNoisePlan,
  s3BucketCreatePlan,
  withDiagnosticsPlan,
} from '@/lib/fixtures/plan-diffs/pulumi'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { PulumiDiff } from './PulumiDiff'

export default {
  title: 'lite/organisms/PulumiDiff',
}

const Frame = ({
  plan,
}: {
  plan: Parameters<typeof PulumiDiff>[0]['plan']
}) => (
  <div className="max-w-6xl p-8">
    <PulumiDiff plan={plan} />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="PulumiDiff"
    tier="organism"
    summary="A Pulumi preview normalized into searchable resource changes and diagnostics."
    use={[
      'Review Pulumi create, update, replace, delete, read, and unchanged resources.',
      'Keep preview diagnostics visible outside the resource section list.',
    ]}
    avoid={[
      'Do not render a Pulumi plan graph from this component.',
      'Do not pass pre-rendered diff lines.',
    ]}
    rules={[
      'Resource inputs serialize deterministically as JSON.',
      'Read and unchanged resources load filtered out and render explanatory notes.',
      'Every legacy Pulumi story uses the exact same shared fixture.',
    ]}
    props={[
      {
        name: 'plan',
        type: 'TPulumiPlan',
        description: 'Raw Pulumi preview response.',
      },
      {
        name: 'defaultOpen',
        type: 'boolean',
        default: 'true',
        description: 'Starting disclosure state for every resource.',
      },
    ]}
  />
)

export const NoPlan = () => <Frame plan={undefined} />

export const S3BucketCreate = () => <Frame plan={s3BucketCreatePlan} />

export const ECSServiceUpdate = () => <Frame plan={ecsServiceUpdatePlan} />

export const DatabaseReplace = () => <Frame plan={databaseReplacePlan} />

export const MixedInfraChanges = () => <Frame plan={mixedInfraChangesPlan} />

export const LargeEKSClusterDeploy = () => (
  <Frame plan={largeEksClusterDeployPlan} />
)

export const KubernetesAppMigration = () => (
  <Frame plan={kubernetesAppMigrationPlan} />
)

export const AzureCosmeticUpdates = () => (
  <Frame plan={azureCosmeticUpdatesPlan} />
)

export const RBACArrayNoise = () => <Frame plan={rbacArrayNoisePlan} />

export const LongPolicyAndConfigValues = () => (
  <Frame plan={longPolicyAndConfigValuesPlan} />
)

export const WithDiagnostics = () => <Frame plan={withDiagnosticsPlan} />
