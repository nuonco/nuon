export default {
  title: 'Diffs/HelmDiff',
}

import type { THelmPlan } from '@/types'
import { HelmDiff } from './HelmDiff'
import {
  certManagerInstallPlan,
  largeDeploymentSingleChangePlan,
  longAnnotationsAndEnvVarsPlan,
  mixedHelmPlan,
  nginxIngressUpgradePlan,
  singleImageTagChangePlan,
  vmagentSingleRemovalPlan,
} from '@/lib/fixtures/plan-diffs/helm'

const plan = (fixture: unknown) => fixture as THelmPlan

export const Mixed = () => (
  <div className="p-4">
    <HelmDiff plan={plan(mixedHelmPlan)} defaultOpen />
  </div>
)

export const IngressUpgrade = () => (
  <div className="p-4">
    <HelmDiff plan={plan(nginxIngressUpgradePlan)} />
  </div>
)

export const FreshInstall = () => (
  <div className="p-4">
    <HelmDiff plan={plan(certManagerInstallPlan)} />
  </div>
)

export const SingleImageTagChange = () => (
  <div className="p-4">
    <HelmDiff plan={plan(singleImageTagChangePlan)} defaultOpen />
  </div>
)

export const LongAnnotationsAndEnvVars = () => (
  <div className="p-4">
    <HelmDiff plan={plan(longAnnotationsAndEnvVarsPlan)} defaultOpen />
  </div>
)

export const LargeManifestSingleChange = () => (
  <div className="p-4">
    <HelmDiff plan={plan(largeDeploymentSingleChangePlan)} defaultOpen />
  </div>
)

export const SingleRemoval = () => (
  <div className="p-4">
    <HelmDiff plan={plan(vmagentSingleRemovalPlan)} defaultOpen />
  </div>
)

export const NoPlan = () => (
  <div className="p-4">
    <HelmDiff />
  </div>
)
