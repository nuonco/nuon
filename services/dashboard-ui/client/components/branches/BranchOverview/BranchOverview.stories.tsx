export default {
  title: 'Features / Branches / Branch overview',
  fullBleed: true,
}

import { useEffect, useState, type ComponentProps } from 'react'
import { BranchRunApproval } from '@/components/branches/BranchRunApproval/BranchRunApproval'
import { PlanDiffPanel } from '@/components/branches/BranchRunApproval/PlanDiffPanel'
import {
  TemplateChangesButton,
  TemplateChangesContent,
} from '@/components/branches/ConfigChanges'
import type { DiffSectionData } from '@/components/approvals/plan-diffs/app-config/AppConfigDiff'
import { GroupActionButton } from '@/components/branches/WorkflowStepDetail/steps/PlanGroupStep/GroupApprovalActions'
import { AppContext } from '@/providers/app-provider'
import { SurfacesProvider } from '@/providers/surfaces-provider'
import type { TInstallWorkflowStep } from '@/types'
import { BuildPanel } from '@/components/builds/BuildPanel'
import { CancelWorkflowButton } from '@/components/workflows/CancelWorkflow'
import type { TWorkflow } from '@/types'
import { InstallDeploymentPanel } from '@/components/installs/InstallDeploymentPanel'
import { useSurfaces } from '@/hooks/use-surfaces'
import { BranchOverview } from './BranchOverview'
import type { TOverviewRollout } from './overview-shared'
import type { TGroupApprovedVersion } from './group-approved-version'
import { buildOverviewLoadingStages } from './overview-loading'
import type { TTrackGroup } from './RolloutTrack'

const storyJson = (body: unknown) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'content-type': 'application/json' },
  })

const storyRequestUrl = (input: RequestInfo | URL) =>
  typeof input === 'string'
    ? input
    : input instanceof Request
      ? input.url
      : String(input)

type TStoryOutcome =
  | 'completed'
  | 'in-progress'
  | 'awaiting'
  | 'error'
  | 'cancelled'

const storyOutcome = (workflowId: string): TStoryOutcome => {
  if (workflowId === 'wf_delta' || workflowId.endsWith('progress')) return 'in-progress'
  if (workflowId.endsWith('awaiting')) return 'awaiting'
  if (workflowId.endsWith('error')) return 'error'
  if (workflowId.endsWith('cancelled')) return 'cancelled'
  return 'completed'
}

const storyStep = (
  workflowId: string,
  name: string,
  status: string,
  extra?: Record<string, unknown>
) => {
  const component = name === 'plan api' || name === 'apply api'
  const finished =
    status === 'success' || status === 'error' || status === 'cancelled'
  return {
    id: `${workflowId}-${name.replace(/\s+/g, '-')}`,
    name,
    execution_type: 'system',
    group_idx: 0,
    group_retry_idx: 0,
    finished,
    execution_time: finished ? 45_000_000_000 : undefined,
    component_name: component ? 'api' : undefined,
    metadata: component ? { component_name: 'api' } : undefined,
    status: { status, history: [] },
    ...extra,
  }
}

const storyWorkflow = (workflowId: string) => {
  const outcome = storyOutcome(workflowId)
  const sync = storyStep(workflowId, 'sync app config', 'success', {
    group_idx: 0,
  })
  const plan = (status: string, extra?: Record<string, unknown>) =>
    storyStep(workflowId, 'plan api', status, { group_idx: 1, ...extra })
  const apply = (status: string, extra?: Record<string, unknown>) =>
    storyStep(workflowId, 'apply api', status, { group_idx: 2, ...extra })
  const sets = {
    completed: {
      status: 'success',
      description: 'Deployed the api component',
      steps: [sync, plan('success'), apply('success')],
    },
    'in-progress': {
      status: 'in-progress',
      description: 'Deploying the api component',
      steps: [
        sync,
        plan('success'),
        apply('in-progress', {
          finished: false,
          execution_time: undefined,
          started_at: '2026-10-15T14:12:00Z',
        }),
      ],
    },
    awaiting: {
      status: 'approval-awaiting',
      description: 'Waiting for plan approval',
      steps: [
        sync,
        plan('approval-awaiting', {
          execution_type: 'approval',
          finished: false,
          execution_time: undefined,
          approval: { id: `${workflowId}-approval`, type: 'terraform_plan' },
        }),
        apply('pending', { finished: false, execution_time: undefined }),
      ],
    },
    error: {
      status: 'error',
      description: 'Deploy failed',
      steps: [sync, plan('success'), apply('error')],
    },
    cancelled: {
      status: 'cancelled',
      description: 'Deployment cancelled',
      steps: [sync, plan('success'), apply('cancelled')],
    },
  } satisfies Record<
    TStoryOutcome,
    { status: string; description: string; steps: unknown[] }
  >
  const set = sets[outcome]
  return {
    id: workflowId,
    type: 'component_deploy',
    approval_option: 'prompt',
    created_at: '2026-10-15T14:12:00Z',
    finished: outcome === 'completed' || outcome === 'error' || outcome === 'cancelled',
    status: { status: set.status, status_human_description: set.description },
    steps: set.steps,
  }
}

const storyDeploymentFetch = (fallback: typeof fetch): typeof fetch =>
  (async (input: RequestInfo | URL, init?: RequestInit) => {
    const method = init?.method ?? (input instanceof Request ? input.method : 'GET')
    const url = storyRequestUrl(input)
    if (method !== 'GET') return fallback(input, init)

    const install = url.match(/\/v1\/installs\/([^/?]+)$/)
    if (install) {
      return storyJson({
        id: install[1],
        name: install[1],
        org_id: 'org-mock-001',
        app_id: 'app-acme',
      })
    }

    const deployment = url.match(/\/v1\/installs\/([^/?]+)\/deployments\/([^/?]+)$/)
    if (deployment) {
      const workflow = storyWorkflow(deployment[2])
      return storyJson({
        id: deployment[2],
        type: 'component_deploy',
        status: workflow.status.status,
        activity: workflow.status.status_human_description,
        created_at: '2026-10-15T14:12:00Z',
        title: 'Deploy acme',
        summary: 'Apply the latest app template to this install.',
        affected_resources: { stack: false, components: ['api'], images: [] },
        change_groups: [],
      })
    }

    if (/\/v1\/installs\/[^/]+\/updates(?:\?|$)/.test(url)) {
      return storyJson({ updates: [], page: 0, limit: 1, has_more: false })
    }

    const steps = url.match(/\/v1\/workflows\/([^/?]+)\/steps(?:\?|$)/)
    if (steps) return storyJson(storyWorkflow(steps[1]).steps)

    const workflow = url.match(/\/v1\/workflows\/([^/?]+)$/)
    if (workflow) {
      return storyJson(storyWorkflow(workflow[1]))
    }

    if (/\/v1\/apps\/[^/]+\/labels$/.test(url)) {
      return storyJson({ labels: [], label_colors: {}, default_colors: [] })
    }

    const componentBuild = url.match(
      /\/v1\/components\/([^/]+)\/builds\/([^/?]+)$/
    )
    if (componentBuild) {
      const componentId = componentBuild[1]
      const name = componentId === 'cmp-cache' ? 'cache' : 'api'
      return storyJson({
        id: componentBuild[2],
        org_id: 'org-mock-001',
        component_id: componentId,
        component_name: name,
        created_at: '2026-10-15T14:00:00Z',
        updated_at: '2026-10-15T14:05:00Z',
        status_description: 'Building the component image.',
        status_v2: {
          status: componentId === 'cmp-cache' ? 'pending' : 'in-progress',
          status_human_description:
            componentId === 'cmp-cache' ? 'Waiting to start' : 'Building image',
        },
        created_by: { email: 'jane@example.com' },
        vcs_connection_commit: {
          sha: 'a1b2c3d4e5f67890',
          message: 'Add cache component',
          author_name: 'Jane Doe',
          created_at: '2026-10-15T13:40:00Z',
        },
      })
    }

    const component = url.match(/\/v1\/components\/([^/?]+)$/)
    if (component) {
      const componentId = component[1]
      return storyJson({
        id: componentId,
        name: componentId === 'cmp-cache' ? 'cache' : 'api',
        type: componentId === 'cmp-cache' ? 'helm_chart' : 'docker_build',
        app_id: 'app-acme',
      })
    }

    return fallback(input, init)
  }) as typeof fetch

const storyApp = {
  app: { id: 'app-acme', name: 'acme-app', org_id: 'org-mock-001' } as never,
  labelColors: {},
  refresh: () => {},
}

const OverviewBody = (props: ComponentProps<typeof BranchOverview>) => {
  const { addPanel } = useSurfaces()
  useEffect(() => {
    const original = window.fetch.bind(window)
    window.fetch = storyDeploymentFetch(original)
    return () => {
      window.fetch = original
    }
  }, [])
  return (
    <BranchOverview
      {...props}
      onSelectInstall={(install) => {
        if (!install.workflowId) return
        addPanel(
          <InstallDeploymentPanel
            orgId="org-acme"
            appId="app-acme"
            installId={install.id}
            workflowId={install.workflowId}
            title={install.name}
            initialTab="workflow"
          />
        )
      }}
      cancelRollout={
        <CancelWorkflowButton
          size="sm"
          workflow={
            {
              id: 'wf_184',
              name: 'Add cache component',
              type: 'app_branches_manual_update',
              status: { status: 'in-progress' },
            } as unknown as TWorkflow
          }
        >
          Cancel rollout
        </CancelWorkflowButton>
      }
      onSelectBuild={(build) => {
        if (!build.componentId) return
        addPanel(
          <BuildPanel
            componentId={build.componentId}
            buildId={build.id}
            title={build.name}
          />
        )
      }}
    />
  )
}

const Overview = (props: ComponentProps<typeof BranchOverview>) => (
  <AppContext.Provider value={storyApp}>
    <SurfacesProvider>
      <OverviewBody {...props} />
    </SurfacesProvider>
  </AppContext.Provider>
)

const groups: TTrackGroup[] = [
  {
    id: 'canary',
    name: 'Canary',
    status: 'success',
    startedAt: '2026-10-15T14:00:00Z',
    installs: [
      {
        id: 'ins_alpha',
        name: 'acme-alpha',
        status: 'success',
        overviewHref: '/org-acme/installs/ins_alpha',
        workflowId: 'wf_alpha',
      },
      {
        id: 'ins_bravo',
        name: 'acme-bravo',
        status: 'success',
        overviewHref: '/org-acme/installs/ins_bravo',
        workflowId: 'wf_bravo',
      },
    ],
  },
  {
    id: 'primary',
    name: 'Primary region',
    status: 'in-progress',
    maxParallel: 1,
    startedAt: '2026-10-15T14:12:00Z',
    installs: [
      {
        id: 'ins_charlie',
        name: 'acme-charlie',
        status: 'success',
        overviewHref: '/org-acme/installs/ins_charlie',
        workflowId: 'wf_charlie',
      },
      {
        id: 'ins_delta',
        name: 'acme-delta',
        status: 'in-progress',
        overviewHref: '/org-acme/installs/ins_delta',
        workflowId: 'wf_delta',
      },
      {
        id: 'ins_echo',
        name: 'acme-echo',
        status: 'pending',
        overviewHref: '/org-acme/installs/ins_echo',
      },
    ],
  },
  {
    id: 'remaining',
    name: 'Remaining',
    status: 'pending',
    plannedCount: 4,
    installs: [],
  },
]

const rollout: TOverviewRollout = {
  id: 'wf_184',
  href: '#run',
  source: {
    kind: 'pull-request',
    number: 482,
    url: 'https://github.com/acme/platform/pull/482',
    baseBranch: 'main',
  },
  title: 'Add cache component',
  sha: 'a1b2c3d4e5f6',
  author: 'jane@example.com',
  status: 'in-progress',
  commit: {
    message: 'Add cache component',
    author: 'jane@example.com',
    sha: 'a1b2c3d4e5f6',
    shaUrl: 'https://github.com/acme/platform/commit/a1b2c3d4e5f6',
    createdAt: '2026-10-15T13:40:00Z',
  },
}

const changeSections: DiffSectionData[] = [
  {
    name: 'Components',
    sectionKey: 'components',
    additions: 1,
    removals: 0,
    changed: 1,
    grouped: true,
    fields: [],
    entities: [
      {
        name: 'cache',
        op: 'add',
        componentType: 'helm_chart',
        fields: [
          { key: 'chart_name', op: 'add', diff: "'cache'" },
          { key: 'namespace', op: 'add', diff: "'acme'" },
          { key: 'values.replicaCount', op: 'add', diff: "'2'" },
        ],
      },
      {
        name: 'api',
        op: 'change',
        componentType: 'helm_chart',
        fields: [
          { key: 'values.replicaCount', op: 'change', diff: "'1' -> '3'" },
        ],
      },
    ],
  },
  {
    name: 'Runner',
    sectionKey: 'runner',
    additions: 0,
    removals: 0,
    changed: 1,
    grouped: false,
    entities: [],
    fields: [
      { key: 'image.tag', op: 'change', diff: "'1.8.2' -> '1.9.0'" },
    ],
  },
]

const changeProps = {
  sections: changeSections,
  versionLabel: 'v6 → v7',
  previousSha: '9f8e7d6c5b4a',
  sha: rollout.sha,
  message: rollout.commit?.message,
  author: rollout.commit?.author,
  createdAt: rollout.commit?.createdAt,
  source: rollout.source,
}

const changes = <TemplateChangesButton {...changeProps} />

const completedBuilds = [
  {
    id: 'build-api',
    name: 'api',
    status: 'success',
    changeReason: 'source_changed',
    componentId: 'cmp-api',
    kind: 'component' as const,
    href: '/org-acme/apps/app-acme/components/cmp-api/builds/build-api',
  },
  {
    id: 'build-cache',
    name: 'cache',
    status: 'success',
    changeReason: 'config_changed',
    componentId: 'cmp-cache',
    kind: 'component' as const,
    href: '/org-acme/apps/app-acme/components/cmp-cache/builds/build-cache',
  },
]

const props = {
  hasPlan: true,
  rollout,
  changes,
  configChanges: <TemplateChangesContent {...changeProps} />,
  configChangesKey: 'v6-v7-cache',
  buildRows: completedBuilds,
  groups,
  versionLabel: 'v6 → v7',
  loadingStages: buildOverviewLoadingStages({
    steps: [
      { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
      {
        id: 'config',
        name: 'sync app config',
        status: { status: 'success' },
      },
      {
        id: 'build',
        name: 'build components',
        status: { status: 'success' },
      },
    ],
    sha: rollout.sha,
    groupStatuses: groups.map((group) => group.status),
  }),
  rolloutHref: '#rollout',
  groupHref: (id: string) => `#rollout/groups/${id}`,
  approvedVersions: {
    canary: {
      version: 12,
      source: { kind: 'tag', tag: 'v1.4.0' },
      sha: 'c0ffee1234567890',
      message: 'Cut the v1.4.0 release',
      author: 'Avery Chen',
      createdAt: '2026-10-14T16:00:00Z',
    },
    'primary region': {
      version: 11,
      source: {
        kind: 'pull-request',
        number: 128,
        url: 'https://github.com/acme/platform/pull/128',
      },
      sha: 'a1b2c3d4e5f6',
      message: 'Add cache component\n\nRolls the cache out to the primary region.',
      author: 'jane@example.com',
      createdAt: '2026-10-15T13:40:00Z',
    },
  } satisfies Record<string, TGroupApprovedVersion>,
}

export const RollingOut = () => <Overview {...props} />
RollingOut.storyName = 'Rolling out'

const outcomeInstalls: TTrackGroup[] = [
  {
    id: 'outcomes',
    name: 'Primary region',
    status: 'in-progress',
    startedAt: '2026-10-15T14:12:00Z',
    installs: [
      {
        id: 'ins_done',
        name: 'acme-done',
        status: 'success',
        overviewHref: '/org-acme/installs/ins_done',
        workflowId: 'wf_completed',
      },
      {
        id: 'ins_live',
        name: 'acme-live',
        status: 'in-progress',
        overviewHref: '/org-acme/installs/ins_live',
        workflowId: 'wf_progress',
      },
      {
        id: 'ins_hold',
        name: 'acme-hold',
        status: 'approval-awaiting',
        overviewHref: '/org-acme/installs/ins_hold',
        workflowId: 'wf_awaiting',
      },
      {
        id: 'ins_bad',
        name: 'acme-bad',
        status: 'error',
        overviewHref: '/org-acme/installs/ins_bad',
        workflowId: 'wf_error',
      },
      {
        id: 'ins_stop',
        name: 'acme-stop',
        status: 'cancelled',
        overviewHref: '/org-acme/installs/ins_stop',
        workflowId: 'wf_cancelled',
      },
    ],
  },
]

export const InstallOutcomes = () => (
  <Overview {...props} groups={outcomeInstalls} />
)
InstallOutcomes.storyName = 'Install outcomes'

const fleetInstall = (
  group: string,
  index: number,
  status: string,
  workflowSuffix = 'completed'
) => {
  const id = `ins_${group}_${index}`
  return {
    id,
    name: `acme-${group}-${index}`,
    status,
    overviewHref: `/org-acme/installs/${id}`,
    ...(status === 'pending'
      ? {}
      : { workflowId: `wf_${group}_${index}_${workflowSuffix}` }),
  }
}

const fleetGroups: TTrackGroup[] = [
  {
    id: 'canary',
    name: 'Canary',
    status: 'success',
    startedAt: '2026-10-15T14:00:00Z',
    installs: [1, 2, 3, 4, 5, 6].map((index) =>
      fleetInstall('canary', index, 'success')
    ),
  },
  {
    id: 'staging',
    name: 'Staging',
    status: 'success',
    startedAt: '2026-10-15T14:20:00Z',
    installs: [1, 2, 3, 4, 5, 6, 7, 8].map((index) =>
      fleetInstall('staging', index, 'success')
    ),
  },
  {
    id: 'us-east',
    name: 'US east',
    status: 'success',
    startedAt: '2026-10-15T14:40:00Z',
    installs: [1, 2, 3, 4, 5, 6, 7].map((index) =>
      fleetInstall('east', index, 'success')
    ),
  },
  {
    id: 'production',
    name: 'Production',
    status: 'in-progress',
    startedAt: '2026-10-15T15:05:00Z',
    installs: [
      fleetInstall('prod', 1, 'success'),
      fleetInstall('prod', 2, 'success'),
      fleetInstall('prod', 3, 'success'),
      fleetInstall('prod', 4, 'in-progress', 'progress'),
      fleetInstall('prod', 5, 'in-progress', 'progress'),
      fleetInstall('prod', 6, 'approval-awaiting', 'awaiting'),
      fleetInstall('prod', 7, 'error', 'error'),
      fleetInstall('prod', 8, 'cancelled', 'cancelled'),
      fleetInstall('prod', 9, 'pending'),
      fleetInstall('prod', 10, 'pending'),
    ],
  },
]

const fleetVersion = (
  version: number,
  message: string
): TGroupApprovedVersion => ({
  version,
  source: {
    kind: 'pull-request',
    number: 482,
    url: 'https://github.com/acme/platform/pull/482',
  },
  sha: 'c0ffee1234567890',
  message,
  author: 'Avery Chen',
  createdAt: '2026-10-14T16:00:00Z',
})

export const FourGroups = () => (
  <Overview
    {...props}
    groups={fleetGroups}
    loadingStages={buildOverviewLoadingStages({
      steps: [
        { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
        { id: 'config', name: 'sync app config', status: { status: 'success' } },
        {
          id: 'build',
          name: 'build components',
          status: { status: 'success' },
        },
      ],
      sha: rollout.sha,
      groupStatuses: fleetGroups.map((group) => group.status),
    })}
    approvedVersions={{
      canary: fleetVersion(12, 'Cut the v1.4.0 release'),
      staging: fleetVersion(11, 'Roll the cache out to staging'),
      'us east': fleetVersion(11, 'Roll the cache out to US east'),
      production: fleetVersion(10, 'Roll the cache out to production'),
    }}
  />
)
FourGroups.storyName = 'Four groups'

const planField = (key: string, op: string, diff: string) => ({
  key,
  diff: { op, diff },
})

const primaryPlanDiff = {
  key: 'app',
  children: [
    {
      key: 'components',
      children: [
        {
          key: 'cache',
          children: [
            planField('type', 'add', "'' -> 'helm_chart'"),
            planField('chart_name', 'add', "'cache'"),
            planField('namespace', 'add', "'acme'"),
          ],
        },
        {
          key: 'api',
          children: [
            planField('type', 'change', "'helm_chart' -> 'helm_chart'"),
            planField('values.replicaCount', 'change', "'1' -> '3'"),
          ],
        },
      ],
    },
    {
      key: 'runner',
      children: [planField('image.tag', 'change', "'1.8.2' -> '1.9.0'")],
    },
  ],
}

const primaryPlanInstalls = [
  ['ins_charlie', 'acme-charlie'],
  ['ins_delta', 'acme-delta'],
  ['ins_echo', 'acme-echo'],
] as const

const primaryPlanContents = {
  install_group: 'Primary region',
  installs: primaryPlanInstalls.map(([install_id, install_name]) => ({
    install_id,
    install_name,
    install_labels: { env: 'prod' },
    old_app_config_id: 'cfg-v6',
    new_app_config_id: 'cfg-v7',
  })),
}

const primaryPlanFacts = Object.fromEntries(
  primaryPlanInstalls.map(([installId]) => [
    installId,
    { status: 'pending', region: 'us-east-1', labels: { env: 'prod' } },
  ])
)

const primaryPlanStep = {
  id: 'step-plan-primary',
  name: 'plan install group: Primary region',
  execution_type: 'approval',
  install_workflow_id: 'wf_184',
  status: { status: 'approval-awaiting' },
  approval: { id: 'apr-primary', type: 'app_branch_plan' },
} as unknown as TInstallWorkflowStep

const PrimaryPlanApproval = () => {
  const { addPanel } = useSurfaces()
  const target = {
    orgId: 'org-acme',
    workflowId: 'wf_184',
    stepId: primaryPlanStep.id!,
    approvalId: 'apr-primary',
    groupName: 'Primary region',
  }
  return (
    <BranchRunApproval
      items={[
        {
          key: primaryPlanStep.id!,
          groupName: 'Primary region',
          onReview: () =>
            addPanel(
              <PlanDiffPanel
                step={primaryPlanStep}
                workflowId="wf_184"
                installFacts={primaryPlanFacts}
              />,
              `plan-diff-${primaryPlanStep.id}`
            ),
          actions: <GroupActionButton action="approve" target={target} />,
        },
      ]}
    />
  )
}

export const AwaitingApproval = () => (
  <AppContext.Provider
    value={{
      app: { id: 'app-acme', name: 'acme-app' } as never,
      labelColors: {},
      refresh: () => {},
    }}
  >
    <SurfacesProvider>
      <AwaitingApprovalOverview />
    </SurfacesProvider>
  </AppContext.Provider>
)
AwaitingApproval.storyName = 'Awaiting approval'

const awaitingPlanGroups = groups.map((group) =>
  group.id === 'primary'
    ? {
        ...group,
        status: 'approval-awaiting',
        installs: group.installs.map(
          ({ workflowId: _workflowId, ...install }) => ({
            ...install,
            status: 'pending',
          })
        ),
      }
    : group
)

const AwaitingApprovalOverview = () => {
  const [approved, setApproved] = useState(false)
  useEffect(() => {
    const original = window.fetch.bind(window)
    const onApproved = () => setApproved(true)
    window.addEventListener('story-plan-approved', onApproved)
    window.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      const url =
        typeof input === 'string'
          ? input
          : input instanceof Request
            ? input.url
            : String(input)
      const method =
        init?.method ?? (input instanceof Request ? input.method : 'GET')
      if (
        method === 'POST' &&
        url.includes('/approvals/') &&
        url.endsWith('/response')
      ) {
        window.dispatchEvent(new Event('story-plan-approved'))
        return new Response(JSON.stringify({ id: 'apr-response' }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        })
      }
      if (method === 'GET' && url.includes('/approvals/') && url.endsWith('/contents')) {
        return storyJson(primaryPlanContents)
      }
      const config = url.match(/\/v1\/apps\/[^/]+\/configs\/(cfg-v\d+)/)
      if (method === 'GET' && config) {
        return storyJson({
          id: config[1],
          version: Number(config[1].replace('cfg-v', '')),
        })
      }
      if (
        method === 'GET' &&
        /\/v1\/installs\/[^/]+\/app-configs\/[^/]+\/diff/.test(url)
      ) {
        return storyJson({
          config_id: 'cfg-v7',
          old_config_id: 'cfg-v6',
          diff: primaryPlanDiff,
          summary: {
            has_changed: true,
            added: 1,
            removed: 0,
            changed: 2,
            unchanged: 0,
          },
          changed: 'components,runner',
        })
      }
      return original(input as RequestInfo, init)
    }) as typeof fetch
    return () => {
      window.fetch = original
      window.removeEventListener('story-plan-approved', onApproved)
    }
  }, [])

  return (
    <Overview
      {...props}
      groups={
        approved
          ? awaitingPlanGroups.map((group) =>
              group.id === 'primary'
                ? { ...group, status: 'in-progress' }
                : group
            )
          : awaitingPlanGroups
      }
      approvals={
        approved
          ? []
          : [
              {
                groupName: 'Primary region',
                banner: <PrimaryPlanApproval />,
              },
            ]
      }
    />
  )
}

const approveAllGroups = groups.map((group) =>
  group.id === 'primary'
    ? {
        ...group,
        status: 'in-progress',
        installs: [
          group.installs[0],
          { ...group.installs[1], status: 'approval-awaiting' },
          {
            ...group.installs[2],
            status: 'approval-awaiting',
            workflowId: 'wf_echo',
          },
        ],
      }
    : group
)

const continueGroup = (plan: TTrackGroup[], groupId: string) =>
  plan.map((group) =>
    group.id === groupId
      ? {
          ...group,
          installs: group.installs.map((install) =>
            install.status === 'approval-awaiting'
              ? { ...install, status: 'in-progress' }
              : install
          ),
        }
      : group
  )

export const ApproveAll = () => {
  const [plan, setPlan] = useState(approveAllGroups)
  useEffect(() => {
    const original = window.fetch.bind(window)
    const approved: string[] = []
    window.fetch = (async (input: RequestInfo | URL, init?: RequestInit) => {
      const url =
        typeof input === 'string'
          ? input
          : input instanceof Request
            ? input.url
            : String(input)
      if (init?.method === 'PATCH' && url.includes('/v1/workflows/')) {
        const body = init.body ? JSON.parse(String(init.body)) : {}
        if (body.approval_option !== 'approve-all') {
          return new Response(JSON.stringify({ error: 'Expected approve-all' }), {
            status: 400,
            headers: { 'content-type': 'application/json' },
          })
        }
        approved.push(url.split('/').pop() ?? '')
        ;(window as unknown as { __approvedWorkflows?: string[] }).__approvedWorkflows =
          approved
        return new Response(JSON.stringify({ id: approved.at(-1) }), {
          status: 200,
          headers: { 'content-type': 'application/json' },
        })
      }
      return original(input as RequestInfo, init)
    }) as typeof fetch
    return () => {
      window.fetch = original
    }
  }, [])

  return (
    <Overview
      {...props}
      groups={plan}
      onApproveAll={(groupId) => setPlan((current) => continueGroup(current, groupId))}
    />
  )
}
ApproveAll.storyName = 'Approve all'

const longGroups: TTrackGroup[] = Array.from({ length: 16 }, (_, index) => {
  const current = index === 8
  const done = index < 8
  return {
    id: `d${index}`,
    name: `d${index}`,
    status: done ? 'success' : current ? 'in-progress' : 'pending',
    startedAt: done || current ? '2026-10-15T14:00:00Z' : undefined,
    installs: [
      {
        id: `ins-d${index}`,
        name: `acme-d${index}`,
        status: done ? 'success' : current ? 'in-progress' : 'pending',
        overviewHref: `/org-acme/installs/ins-d${index}`,
        workflowId: done || current ? `wf-d${index}` : undefined,
      },
    ],
  }
})

export const ManyGroups = () => (
  <Overview {...props} groups={longGroups} />
)
ManyGroups.storyName = 'Many groups'

const beforeRolloutGroups: TTrackGroup[] = groups.map((group) => ({
  ...group,
  status: 'pending',
  startedAt: undefined,
  installs: group.installs.map(({ workflowId: _workflowId, ...install }) => ({
    ...install,
    status: 'pending',
  })),
}))

export const Building = () => (
  <Overview
    {...props}
    groups={beforeRolloutGroups}
    buildRows={[
      { ...completedBuilds[0], status: 'in-progress' },
      { ...completedBuilds[1], status: 'pending' },
    ]}
    loadingStages={buildOverviewLoadingStages({
      steps: [
        { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
        { id: 'config', name: 'sync app config', status: { status: 'success' } },
        {
          id: 'build',
          name: 'build components',
          status: { status: 'in-progress' },
        },
      ],
      sha: rollout.sha,
      groupStatuses: beforeRolloutGroups.map((group) => group.status),
    })}
  />
)
Building.storyName = 'Building'

export const BuildComplete = () => (
  <Overview
    {...props}
    groups={beforeRolloutGroups}
    loadingStages={buildOverviewLoadingStages({
      steps: [
        { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
        { id: 'config', name: 'sync app config', status: { status: 'success' } },
        {
          id: 'build',
          name: 'build components',
          status: { status: 'success' },
        },
      ],
      sha: rollout.sha,
      groupStatuses: beforeRolloutGroups.map((group) => group.status),
    })}
  />
)
BuildComplete.storyName = 'Build complete'

const pendingGroups = beforeRolloutGroups.map((group) => group.status)

const pendingChanges = (
  <TemplateChangesButton {...changeProps} sections={[]} isPending />
)
const pendingChangeContent = (
  <TemplateChangesContent {...changeProps} sections={[]} isPending />
)

const stagesFor = (steps: TInstallWorkflowStep[], sha?: string) =>
  buildOverviewLoadingStages({
    steps,
    sha,
    groupStatuses: pendingGroups,
  })

const step = (
  id: string,
  name: string,
  status: string
): TInstallWorkflowStep =>
  ({ id, name, status: { status } }) as TInstallWorkflowStep

export const StartingWorkflow = () => (
  <Overview
    {...props}
    groups={beforeRolloutGroups}
    rollout={{ ...rollout, sha: undefined, commit: undefined }}
    changes={pendingChanges}
    configChanges={pendingChangeContent}
    buildRows={[]}
    loadingStages={stagesFor([])}
  />
)
StartingWorkflow.storyName = 'Starting workflow'

export const FetchingCommit = () => (
  <Overview
    {...props}
    groups={beforeRolloutGroups}
    rollout={{ ...rollout, sha: undefined, commit: undefined }}
    changes={pendingChanges}
    configChanges={pendingChangeContent}
    buildRows={[]}
    loadingStages={stagesFor([
      step('fetch', 'fetch commit', 'in-progress'),
    ])}
  />
)
FetchingCommit.storyName = 'Fetching commit'

export const CompilingConfig = () => (
  <Overview
    {...props}
    groups={beforeRolloutGroups}
    changes={pendingChanges}
    configChanges={pendingChangeContent}
    buildRows={[]}
    loadingStages={stagesFor(
      [
        step('fetch', 'fetch commit', 'success'),
        step('config', 'sync app config', 'in-progress'),
      ],
      rollout.sha
    )}
  />
)
CompilingConfig.storyName = 'Compiling config'
