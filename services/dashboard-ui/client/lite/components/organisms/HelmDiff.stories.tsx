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
} from '../../lib/fixtures/plan-diffs/helm'
import { ComponentDocs } from '../__stories__/ComponentDocs'
import { HelmDiff } from './HelmDiff'

export default {
  title: 'lite/organisms/HelmDiff',
}

const Frame = ({ plan }: { plan: Parameters<typeof HelmDiff>[0]['plan'] }) => (
  <div className="max-w-6xl p-8">
    <HelmDiff plan={plan} />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="HelmDiff"
    tier="organism"
    summary="A Helm plan normalized into the shared summary, filter and diff sections."
    use={[
      'Review every Kubernetes resource changed by one Helm operation.',
      'Search resource identity and filter create, update or delete operations.',
    ]}
    avoid={[
      'Do not pass pre-rendered unified diff lines.',
      'Do not put provider-specific controls inside individual sections.',
    ]}
    rules={[
      'Both direct before/after payloads and entry-array payloads normalize to YAML.',
      'The summary always describes the complete plan, even while sections are filtered.',
      'All legacy Helm diff scenarios have a matching lite story.',
    ]}
    props={[
      {
        name: 'plan',
        type: 'THelmPlan',
        description: 'Raw Helm planner response.',
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

export const Default = () => <Frame plan={mixedHelmPlan} />

export const NginxIngressUpgrade = () => (
  <Frame plan={nginxIngressUpgradePlan} />
)

export const CertManagerInstall = () => <Frame plan={certManagerInstallPlan} />

export const PostgresOperatorUpgrade = () => (
  <Frame plan={postgresOperatorUpgradePlan} />
)

export const PrometheusStackChange = () => (
  <Frame plan={prometheusStackChangePlan} />
)

export const RedisClusterRollback = () => (
  <Frame plan={redisClusterRollbackPlan} />
)

export const VmagentSingleRemoval = () => (
  <Frame plan={vmagentSingleRemovalPlan} />
)

export const LongAnnotationsAndEnvVars = () => (
  <Frame plan={longAnnotationsAndEnvVarsPlan} />
)

export const SingleImageTagChange = () => (
  <Frame plan={singleImageTagChangePlan} />
)

export const LargeDeploymentSingleChange = () => (
  <Frame plan={largeDeploymentSingleChangePlan} />
)

export const LargeDeploymentScatteredChanges = () => (
  <Frame plan={largeDeploymentScatteredChangesPlan} />
)
