import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { InstallGroupCard } from './InstallGroupCard'

export default {
  title: 'lite/organisms/InstallGroupCard',
}

const INSTALLS = [
  {
    id: 'inst_alpha',
    name: 'alpha',
    cloud_platform: 'aws',
    aws_account: { region: 'us-west-2' },
    runner_status: 'active',
    sandbox_status: 'active',
    sandbox_health_status: 'healthy',
    composite_component_status: 'healthy',
  },
  {
    id: 'inst_bravo',
    name: 'bravo',
    cloud_platform: 'aws',
    aws_account: { region: 'us-east-1' },
    runner_status: 'active',
    sandbox_status: 'active',
    sandbox_health_status: 'healthy',
    composite_component_status: 'deploying',
  },
  {
    id: 'inst_charlie',
    name: 'charlie',
    cloud_platform: 'gcp',
    gcp_account: { region: 'us-central1' },
    runner_status: 'pending',
    sandbox_status: 'provisioning',
    composite_component_status: 'pending',
  },
]

const STAGE: IDeploymentPlanStage = {
  id: 'grp_core',
  name: 'Core installs',
  order: 0,
  stage: 1,
  membership: 'install_ids',
  installs: INSTALLS,
  totalInstalls: 3,
  maxParallel: 2,
  autoApproveOnPoliciesPassing: true,
  status: 'success',
  completedInstalls: 3,
  failedInstalls: 0,
}

export const Overview = () => (
  <ComponentDocs
    name="InstallGroupCard"
    tier="organism"
    summary="One ordered deployment stage with its membership rule, rollout state, and a bounded install preview."
    use={[
      'Render one resolved deployment-plan stage inside DeploymentPlanStages.',
      'Pass groupHref for the complete, paginated membership panel.',
      'Pass onInstallSelect when install rows should open summary panels.',
    ]}
    avoid={[
      'Do not render every install inline. The card shows at most five.',
      'Do not calculate membership or join group runs in this component.',
      'Do not add approval actions. This card only identifies the waiting state.',
    ]}
    rules={[
      'Run totals take precedence because the resolver supplies what the rollout covered.',
      'Selector groups show their chips and explain that membership is dynamic.',
      'Explicit-list groups do not show the dynamic-membership line.',
      'Empty groups remain visible with a zero count.',
      'Install rows reuse ConfigItem.',
    ]}
    props={[
      {
        name: 'stage',
        type: 'IDeploymentPlanStage',
        description: 'Resolved group membership and optional rollout overlay.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, string>',
        description: 'App label colours keyed by label key.',
      },
      {
        name: 'groupHref',
        type: 'string',
        description: 'Link to the group membership panel.',
      },
      {
        name: 'onInstallSelect',
        type: '(install: IDeploymentPlanInstall) => void',
        description: 'Opens the selected install summary panel.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads values inside the real card structure.',
      },
    ]}
  />
)

export const Default = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard
      stage={STAGE}
      groupHref="?panel=group%3Agrp_core"
      onInstallSelect={() => {}}
    />
  </div>
)

export const SelectorGroup = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard
      stage={{
        ...STAGE,
        id: 'grp_selector',
        name: 'Production',
        membership: 'label_selector',
        selector: {
          match_labels: { env: 'prod', tier: '*' },
          not_match_labels: { region: 'us-east-1' },
        },
      }}
      labelColors={{
        env: '#4cc9f0',
        tier: '#4aa578',
        region: '#8b5cf6',
      }}
      groupHref="?panel=group%3Agrp_selector"
      onInstallSelect={() => {}}
    />
  </div>
)

export const AwaitingApproval = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard
      stage={{ ...STAGE, status: 'approval-awaiting' }}
      groupHref="?panel=group%3Agrp_core"
      onInstallSelect={() => {}}
    />
  </div>
)

export const LargeGroup = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard
      stage={{
        ...STAGE,
        installs: Array.from({ length: 340 }, (_, index) => ({
          id: `inst_${index}`,
          name: `install-${index + 1}`,
        })),
        totalInstalls: 340,
        status: 'in-progress',
        completedInstalls: 128,
        failedInstalls: 2,
      }}
      groupHref="?panel=group%3Agrp_core"
      onInstallSelect={() => {}}
    />
  </div>
)

export const EmptyGroup = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard
      stage={{ ...STAGE, installs: [], totalInstalls: 0, status: undefined }}
    />
  </div>
)

export const Loading = () => (
  <div className="max-w-3xl p-8">
    <InstallGroupCard loading />
  </div>
)
