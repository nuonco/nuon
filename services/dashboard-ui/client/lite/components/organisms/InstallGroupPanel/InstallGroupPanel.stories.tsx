import { useState } from 'react'
import type { IDeploymentPlanStage } from '../../../utils/deployment-plan'
import { ComponentDocs } from '../../__stories__/ComponentDocs'
import { SurfaceStory } from '../../__stories__/SurfaceStory'
import { InstallGroupPanel } from './InstallGroupPanel'

export default {
  title: 'lite/organisms/InstallGroupPanel',
}

const STAGE: IDeploymentPlanStage = {
  id: 'grp_primary',
  name: 'Primary region',
  order: 1,
  stage: 2,
  membership: 'label_selector',
  selector: {
    match_labels: { env: 'prod', region: '*' },
    not_match_labels: { tier: 'preview' },
  },
  installs: Array.from({ length: 45 }, (_, index) => ({
    id: `inst_${index}`,
    name: `install-${index + 1}`,
    cloud_platform: 'aws',
    aws_account: { region: index % 2 ? 'us-east-1' : 'us-west-2' },
    runner_status: 'active',
    sandbox_status: 'active',
    sandbox_health_status: 'healthy',
    composite_component_status: index < 3 ? 'deploying' : 'healthy',
  })),
  totalInstalls: 45,
  completedInstalls: 31,
  failedInstalls: 2,
  status: 'in-progress',
  maxParallel: 4,
  autoApproveOnPoliciesPassing: false,
}

const StatefulPanel = ({
  stage = STAGE,
  loading = false,
  error,
}: {
  stage?: IDeploymentPlanStage
  loading?: boolean
  error?: unknown
}) => {
  const [offset, setOffset] = useState(0)

  return (
    <InstallGroupPanel
      stage={stage}
      run={{
        install_group_id: stage?.id,
        total_installs: stage?.totalInstalls,
        completed_installs: stage?.completedInstalls,
        failed_installs: stage?.failedInstalls,
        status: stage?.status ? { status: stage.status as never } : undefined,
        updated_at: '2026-09-08T13:45:00Z',
      }}
      labelColors={{
        env: '#4cc9f0',
        region: '#8b5cf6',
        tier: '#4aa578',
      }}
      offset={offset}
      pageSize={20}
      onOffsetChange={setOffset}
      onInstallSelect={() => {}}
      loading={loading}
      error={error}
    />
  )
}

const openPanel = (panel: React.ReactElement) => (surfaces: {
  openPanel: (content: React.ReactElement) => string
}) => {
  surfaces.openPanel(panel)
}

export const Overview = () => (
  <ComponentDocs
    name="InstallGroupPanel"
    tier="organism"
    summary="Linkable deployment-group detail with the written rule, rollout state, and paged resolved installs."
    use={[
      'Open from View all installs on an install group card.',
      'Use the container so a direct panel URL resolves current membership.',
    ]}
    avoid={[
      'Do not render the full list inside the stage card.',
      'Do not add approve or deny controls.',
      'Do not treat current selector matches as the rollout total.',
    ]}
    rules={[
      'The membership rule and label selector are shown exactly as configured.',
      'The latest rollout keeps its own total beside the currently resolved members.',
      'The install list pages at twenty rows.',
      'Install rows open the install summary panel.',
    ]}
    props={[
      {
        name: 'stage',
        type: 'IDeploymentPlanStage',
        description: 'Resolved group and rollout overlay.',
      },
      {
        name: 'run',
        type: 'TInstallGroupRun',
        description: 'Latest group run timestamps and counts.',
      },
      {
        name: 'labelColors',
        type: 'Record<string, string>',
        description: 'App label colours keyed by label key.',
      },
      {
        name: 'offset',
        type: 'number',
        description: 'Current install-list offset.',
      },
      {
        name: 'pageSize',
        type: 'number',
        description: 'Install rows shown per page.',
      },
      {
        name: 'onOffsetChange',
        type: '(offset: number) => void',
        description: 'Moves the install-list page.',
      },
      {
        name: 'onInstallSelect',
        type: '(install: IDeploymentPlanInstall) => void',
        description: 'Opens the install summary panel.',
      },
      {
        name: 'loading',
        type: 'boolean',
        default: 'false',
        description: 'Loads fields and install rows within the panel.',
      },
      {
        name: 'error',
        type: 'unknown',
        description: 'Shows the group failed-to-load state.',
      },
    ]}
  />
)

export const Default = () => (
  <SurfaceStory open={openPanel(<StatefulPanel />)} />
)

export const AwaitingApproval = () => (
  <SurfaceStory
    open={openPanel(
      <StatefulPanel stage={{ ...STAGE, status: 'approval-awaiting' }} />
    )}
  />
)

export const Empty = () => (
  <SurfaceStory
    open={openPanel(
      <StatefulPanel stage={{ ...STAGE, installs: [], totalInstalls: 0 }} />
    )}
  />
)

export const Loading = () => (
  <SurfaceStory open={openPanel(<StatefulPanel loading />)} />
)

export const FailedToLoad = () => (
  <SurfaceStory
    open={openPanel(<StatefulPanel error={new Error('failed')} />)}
  />
)
