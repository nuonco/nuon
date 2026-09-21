export default {
  title: 'Diffs/TerraformDiff',
}

import type { TTerraformPlan } from '@/types'
import { TerraformDiff } from './TerraformDiff'
import {
  driftDetectedPlan,
  eksClusterCreatePlan,
  iamRoleWithNestedPolicyPlan,
  mixedWithNoOpPlan,
  rdsReplacePlan,
  securityGroupUpdatePlan,
} from '@/lib/fixtures/plan-diffs/terraform'

const plan = (fixture: unknown) => fixture as TTerraformPlan

export const NestedIAMPolicy = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(iamRoleWithNestedPolicyPlan)} defaultOpen />
  </div>
)

export const ClusterCreate = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(eksClusterCreatePlan)} />
  </div>
)

export const SecurityGroupUpdate = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(securityGroupUpdatePlan)} defaultOpen />
  </div>
)

export const Replace = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(rdsReplacePlan)} defaultOpen />
  </div>
)

export const MixedWithNoOp = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(mixedWithNoOpPlan)} />
  </div>
)

export const DriftDetected = () => (
  <div className="p-4">
    <TerraformDiff plan={plan(driftDetectedPlan)} />
  </div>
)

export const NoPlan = () => (
  <div className="p-4">
    <TerraformDiff />
  </div>
)
