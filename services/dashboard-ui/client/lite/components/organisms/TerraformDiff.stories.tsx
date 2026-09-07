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
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { TerraformDiff } from './TerraformDiff'

export default {
  title: 'lite/organisms/TerraformDiff',
}

const Frame = ({
  plan,
}: {
  plan: Parameters<typeof TerraformDiff>[0]['plan']
}) => (
  <div className="max-w-6xl p-8">
    <TerraformDiff plan={plan} />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="TerraformDiff"
    tier="organism"
    summary="A Terraform plan normalized into drift, resource, and output groups."
    use={[
      'Review resource drift, planned resource changes, and output changes separately.',
      'Search addresses and filter create, update, replace, delete, read, and no-op operations.',
    ]}
    avoid={[
      'Do not pass pre-rendered unified diff lines.',
      'Do not render a plan graph from this component.',
    ]}
    rules={[
      'Before and after values serialize through the shared Terraform serializer.',
      'Cosmetic null/empty drift is omitted from the drift group.',
      'All legacy Terraform diff scenarios have a matching lite story.',
    ]}
    props={[
      {
        name: 'plan',
        type: 'TTerraformPlan',
        description: 'Raw Terraform planner response.',
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

export const WithPlan = () => <Frame plan={withPlan} />

export const IAMRoleWithNestedPolicy = () => (
  <Frame plan={iamRoleWithNestedPolicyPlan} />
)

export const EKSClusterCreate = () => <Frame plan={eksClusterCreatePlan} />

export const SecurityGroupUpdate = () => (
  <Frame plan={securityGroupUpdatePlan} />
)

export const RDSReplace = () => <Frame plan={rdsReplacePlan} />

export const NoOpAndReadResources = () => (
  <Frame plan={noOpAndReadResourcesPlan} />
)

export const ReplaceResources = () => <Frame plan={replaceResourcesPlan} />

export const MixedWithNoOp = () => <Frame plan={mixedWithNoOpPlan} />

export const RBACArrayNoise = () => <Frame plan={rbacArrayNoisePlan} />

export const AzureNoOpWithCosmeticDrift = () => (
  <Frame plan={azureNoOpWithCosmeticDriftPlan} />
)

export const DriftDetected = () => <Frame plan={driftDetectedPlan} />

export const KubectlManifestDeployment = () => (
  <Frame plan={kubectlManifestDeploymentPlan} />
)

export const DriftWithChangesAndOutputs = () => (
  <Frame plan={driftWithChangesAndOutputsPlan} />
)
