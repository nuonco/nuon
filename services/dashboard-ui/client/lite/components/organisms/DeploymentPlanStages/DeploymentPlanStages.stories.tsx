import type { TAppBranch } from '@/types'
import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { DeploymentPlanStages } from './DeploymentPlanStages'

export default {
  title: 'lite/organisms/DeploymentPlanStages',
}

const BRANCH: TAppBranch = {
  id: 'br_main',
  name: 'main',
  latest_run: {
    id: 'run_latest',
    status: 'in-progress',
    head_sha: 'a1b2c3d4e5f6',
  },
}

const stages = (
  statuses: Array<string | undefined> = ['success', 'in-progress', 'pending']
): IDeploymentPlanStage[] => [
  {
    id: 'grp_core',
    name: 'Core',
    order: 0,
    stage: 1,
    membership: 'install_ids',
    installs: [
      { id: 'inst_alpha', name: 'alpha' },
      { id: 'inst_bravo', name: 'bravo' },
    ],
    totalInstalls: 2,
    completedInstalls: statuses[0] ? 2 : undefined,
    failedInstalls: statuses[0] ? 0 : undefined,
    status: statuses[0],
    maxParallel: 2,
    autoApproveOnPoliciesPassing: true,
  },
  {
    id: 'grp_region',
    name: 'Primary region',
    order: 1,
    stage: 2,
    membership: 'label_selector',
    selector: {
      match_labels: { env: 'prod', region: '*' },
      not_match_labels: { tier: 'preview' },
    },
    installs: [
      { id: 'inst_charlie', name: 'charlie' },
      { id: 'inst_delta', name: 'delta' },
      { id: 'inst_echo', name: 'echo' },
    ],
    totalInstalls: 3,
    completedInstalls: statuses[1] ? 1 : undefined,
    failedInstalls: statuses[1] ? 0 : undefined,
    status: statuses[1],
    maxParallel: 1,
    autoApproveOnPoliciesPassing: false,
  },
  {
    id: 'grp_rest',
    name: 'Remaining installs',
    order: 2,
    stage: 3,
    membership: 'all_installs',
    installs: [{ id: 'inst_foxtrot', name: 'foxtrot' }],
    totalInstalls: 1,
    completedInstalls: statuses[2] ? 0 : undefined,
    failedInstalls: statuses[2] ? 0 : undefined,
    status: statuses[2],
  },
]

const groupHref = (stage: IDeploymentPlanStage) =>
  `?panel=group%3A${stage.id}`

const Story = ({
  value = stages(),
  branch = BRANCH,
}: {
  value?: IDeploymentPlanStage[]
  branch?: TAppBranch
}) => (
  <div className="max-w-3xl p-8">
    <DeploymentPlanStages
      branch={branch}
      stages={value}
      labelColors={{
        env: '#4cc9f0',
        region: '#8b5cf6',
        tier: '#4aa578',
      }}
      groupHref={groupHref}
      onInstallSelect={() => {}}
    />
  </div>
)

export const Overview = () => (
  <ComponentDocs
    name="DeploymentPlanStages"
    tier="organism"
    summary="A branch deployment plan as a fixed vertical sequence with the latest rollout laid over its groups."
    use={[
      'Render the branch deployment plan on the branch overview.',
      'Use the container so install membership and group runs resolve outside presentation.',
    ]}
    avoid={[
      'Do not turn this sequence into a graph or horizontal layout.',
      'Do not fetch or resolve group membership in the presentation component.',
      'Do not put the full install list in a stage card.',
    ]}
    rules={[
      'The branch summary identifies the branch, head commit, latest run, and status.',
      'Stages stay vertical at every width and are joined by simple dividers.',
      'Loading uses the real cards with loading primitives.',
      'No plan and failed to load use distinct copy in the same slot.',
    ]}
    props={[
      {
        name: 'branch',
        type: 'TAppBranch',
        description: 'Branch identity and latest rollout summary.',
      },
      {
        name: 'stages',
        type: 'IDeploymentPlanStage[]',
        description: 'Ordered, resolved stages with optional run state.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, string>',
        description: 'App label colours keyed by label key.',
      },
      {
        name: 'groupHref',
        type: '(stage: IDeploymentPlanStage) => string | undefined',
        description: 'Builds the link to each group panel.',
      },
      {
        name: 'onInstallSelect',
        type: '(install: IDeploymentPlanInstall) => void',
        description: 'Opens an install summary panel.',
      },
      {
        name: 'renderStageCard',
        type: '(stage: IDeploymentPlanStage) => ReactNode',
        description:
          'Container-owned stage renderer used to bind linkable panel controls.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads the branch summary and three real stage cards.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Switches the body to the failed-to-load state.',
      },
    ]}
  />
)

export const Default = () => <Story />

export const InProgress = () => <Story value={stages()} />

export const Failed = () => (
  <Story value={stages(['success', 'error', 'pending'])} />
)

export const AwaitingApproval = () => (
  <Story value={stages(['success', 'approval-awaiting', 'pending'])} />
)

export const NoRuns = () => (
  <Story
    branch={{ id: 'br_main', name: 'main' }}
    value={stages([undefined, undefined, undefined])}
  />
)

export const SelectorGroup = () => <Story value={[stages()[1]!]} />

export const LargeGroup = () => (
  <Story
    value={[
      {
        ...stages()[0]!,
        installs: Array.from({ length: 340 }, (_, index) => ({
          id: `inst_${index}`,
          name: `install-${index + 1}`,
        })),
        totalInstalls: 340,
        completedInstalls: 128,
        failedInstalls: 2,
        status: 'in-progress',
      },
    ]}
  />
)

export const EmptyGroup = () => (
  <Story
    value={[
      {
        ...stages()[1]!,
        installs: [],
        totalInstalls: 0,
        completedInstalls: undefined,
        failedInstalls: undefined,
        status: undefined,
      },
    ]}
  />
)

export const SingleGroup = () => <Story value={[stages()[0]!]} />

export const Loading = () => (
  <div className="max-w-3xl p-8">
    <DeploymentPlanStages loading />
  </div>
)

export const Empty = () => <Story value={[]} />

export const FailedToLoad = () => (
  <div className="max-w-3xl p-8">
    <DeploymentPlanStages branch={BRANCH} error={new Error('failed')} />
  </div>
)
