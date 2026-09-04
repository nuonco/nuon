import {
  certManagerInstallPlan,
  largeDeploymentScatteredChangesPlan,
  largeDeploymentSingleChangePlan,
  longAnnotationsAndEnvVarsPlan,
  mixedHelmPlan,
  nginxIngressUpgradePlan,
  postgresOperatorUpgradePlan,
  prometheusStackChangePlan,
  redisClusterRollbackPlan,
  singleImageTagChangePlan,
  vmagentSingleRemovalPlan,
} from '@/lib/fixtures/plan-diffs/helm'
import { HelmDiff } from './HelmDiff'

export default {
  title: 'Approvals/PlanDiffs/HelmDiff',
}

export const Default = () => <HelmDiff plan={mixedHelmPlan} />

export const NginxIngressUpgrade = () => (
  <HelmDiff plan={nginxIngressUpgradePlan} />
)

export const CertManagerInstall = () => (
  <HelmDiff plan={certManagerInstallPlan} />
)

export const PostgresOperatorUpgrade = () => (
  <HelmDiff plan={postgresOperatorUpgradePlan} />
)

export const PrometheusStackChange = () => (
  <HelmDiff plan={prometheusStackChangePlan} />
)

export const RedisClusterRollback = () => (
  <HelmDiff plan={redisClusterRollbackPlan} />
)

export const VmagentSingleRemoval = () => (
  <HelmDiff plan={vmagentSingleRemovalPlan} />
)

export const LongAnnotationsAndEnvVars = () => (
  <HelmDiff plan={longAnnotationsAndEnvVarsPlan} />
)

export const SingleImageTagChange = () => (
  <HelmDiff plan={singleImageTagChangePlan} />
)

export const LargeDeploymentSingleChange = () => (
  <HelmDiff plan={largeDeploymentSingleChangePlan} />
)

export const LargeDeploymentScatteredChanges = () => (
  <HelmDiff plan={largeDeploymentScatteredChangesPlan} />
)
