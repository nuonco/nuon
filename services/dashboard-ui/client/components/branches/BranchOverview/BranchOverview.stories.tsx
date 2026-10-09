export default {
  title: 'Features / Branches / Branch overview',
  fullBleed: true,
}

import { ConfigParseFailure } from '@/components/branches/BranchRunChangesSummary/ConfigParseFailure'
import { TemplateChangesButton } from '@/components/branches/ConfigChanges'
import type { TCompositeError } from '@/types'
import { BranchOverview, type TOverviewRollout } from './BranchOverview'
import { buildOverviewLoadingStages } from './overview-loading'
import type { TTrackGroup } from './RolloutTrack'

const groups: TTrackGroup[] = [
  {
    id: 'canary',
    name: 'Canary',
    status: 'success',
    rules: 'tier=canary · up to 2 at a time · auto-approves when policies pass',
    installs: [
      { id: 'ins_alpha', name: 'alpha', status: 'success' },
      { id: 'ins_bravo', name: 'bravo', status: 'success' },
    ],
  },
  {
    id: 'primary',
    name: 'Primary region',
    status: 'in-progress',
    rules: 'region=us-east-1 · up to 2 at a time · manual approval',
    installs: [
      { id: 'ins_delta', name: 'delta', status: 'in-progress' },
      { id: 'ins_charlie', name: 'charlie', status: 'success' },
      { id: 'ins_echo', name: 'echo', status: 'pending' },
    ],
  },
  {
    id: 'remaining',
    name: 'Remaining',
    status: 'pending',
    plannedCount: 14,
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
    label: 'deploy',
    baseBranch: 'main',
  },
  title: 'Add cache component',
  sha: 'a1b2c3d4e5f6',
  author: 'jane@example.com',
  status: 'in-progress',
  activity: 'delta started deploying',
}

const changes = (
  <TemplateChangesButton
    versionLabel="v13 → v14"
    previousSha="9f8e7d6c5b4a"
    sha="a1b2c3d4e5f6"
    sections={[
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
            fields: [{ key: 'chart_name', op: 'add', diff: "'cache'" }],
          },
          {
            name: 'api',
            op: 'change',
            componentType: 'helm_chart',
            fields: [
              {
                key: 'image.tag',
                op: 'change',
                diff: "'1.4.1' -> '1.4.2'",
              },
            ],
          },
        ],
      },
    ]}
  />
)

export const RollingOut = () => (
  <BranchOverview
    hasPlan
    rollout={rollout}
    changes={changes}
    groups={groups}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
RollingOut.storyName = 'Rolling out'

export const TagPush = () => (
  <BranchOverview
    hasPlan
    rollout={{
      ...rollout,
      source: {
        kind: 'tag',
        tag: 'v1.4.2',
        url: 'https://github.com/acme/platform/releases/tag/v1.4.2',
      },
      title: 'Bump api to 1.4.2',
      status: 'success',
    }}
    changes={changes}
    groups={groups.map((group) => ({
      ...group,
      status: 'success',
      installs: group.installs.map((install) => ({
        ...install,
        status: 'success',
      })),
    }))}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
TagPush.storyName = 'Tag push'

const manualRollout: TOverviewRollout = {
  ...rollout,
  source: { kind: 'manual' },
  title: 'Manual run',
  sha: undefined,
  author: undefined,
  status: 'in-progress',
  activity: 'Executing step fetch commit',
}

export const WaitingForWorkflow = () => (
  <BranchOverview
    hasPlan
    isLoading
    loadingStages={buildOverviewLoadingStages({ steps: [] })}
    groups={[]}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
WaitingForWorkflow.storyName = 'Waiting for workflow'

export const FetchingCommit = () => (
  <BranchOverview
    hasPlan
    rollout={manualRollout}
    changes={changes}
    loadingStages={buildOverviewLoadingStages({
      steps: [
        {
          id: 'fetch',
          name: 'fetch commit',
          status: { status: 'in-progress' },
        },
      ],
    })}
    groups={groups}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
FetchingCommit.storyName = 'Fetching commit'

export const CommitReady = () => (
  <BranchOverview
    hasPlan
    rollout={{
      ...manualRollout,
      sha: 'a1b2c3d4e5f6',
      author: 'jane@example.com',
      commit: {
        message: 'Add cache component',
        author: 'jane@example.com',
        sha: 'a1b2c3d4e5f6',
        shaUrl: 'https://github.com/acme/platform/commit/a1b2c3d4e5f6',
        createdAt: '2026-09-29T18:00:00Z',
      },
    }}
    changes={changes}
    loadingStages={buildOverviewLoadingStages({
      steps: [
        { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
        {
          id: 'config',
          name: 'sync app config',
          status: { status: 'in-progress' },
        },
      ],
      sha: 'a1b2c3d4e5f6',
    })}
    groups={groups}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
CommitReady.storyName = 'Commit ready'

const buildError = {
  message: 'api image build failed',
  severity: 'error',
} as TCompositeError

export const BuildFailed = () => (
  <BranchOverview
    hasPlan
    rollout={{
      ...manualRollout,
      status: 'error',
      sha: 'a1b2c3d4e5f6',
      commit: {
        message: 'Add cache component',
        author: 'jane@example.com',
        sha: 'a1b2c3d4e5f6',
        shaUrl: 'https://github.com/acme/platform/commit/a1b2c3d4e5f6',
      },
    }}
    changes={changes}
    loadingStages={buildOverviewLoadingStages({
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
          status: { status: 'error', composite_error: buildError },
        },
      ],
      sha: 'a1b2c3d4e5f6',
    })}
    compositeError={buildError}
    failedBuilds={[
      {
        id: 'bld_api',
        name: 'api',
        href: '#components/api/builds/bld_api',
      },
    ]}
    groups={groups}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
BuildFailed.storyName = 'Build failed'

export const ConfigUnparseable = () => (
  <BranchOverview
    hasPlan
    showInstalls={false}
    rollout={{
      ...manualRollout,
      status: 'error',
      sha: 'a1b2c3d4e5f6',
      author: 'jane@example.com',
      commit: {
        message: 'Add cache component',
        author: 'jane@example.com',
        sha: 'a1b2c3d4e5f6',
        shaUrl: 'https://github.com/acme/platform/commit/a1b2c3d4e5f6',
      },
    }}
    changes={
      <ConfigParseFailure
        lines={[
          'components/api.toml: unable to parse configuration file: toml: line 4: expected key but found end of file',
          'components/worker.toml: image is required',
        ]}
      />
    }
    loadingStages={buildOverviewLoadingStages({
      steps: [
        { id: 'fetch', name: 'fetch commit', status: { status: 'success' } },
        {
          id: 'config',
          name: 'sync app config',
          status: { status: 'error' },
        },
      ],
      sha: 'a1b2c3d4e5f6',
    })}
    groups={[]}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
ConfigUnparseable.storyName = 'Config could not be parsed'

const longCommit = [
  'feat: map preview install configs onto branch groups (#104)',
  '',
  '* rename the ramp config to Customer A for the preview group',
  '* rename the sony config to Customer B for the push group',
  '* add Customer C install config for the manual group',
  '* add Customer D install config for the tag group',
  '* move prod install configs under install-configs/prod',
  '* add a stage folder and Customer A install config',
  '* pin AWS install configs to their app branches',
].join('\n')

export const LongCommitMessage = () => (
  <BranchOverview
    hasPlan
    rollout={{
      ...manualRollout,
      status: 'error',
      title: longCommit,
      sha: 'e5aef07',
      author: 'casey@example.com',
      commit: {
        message: longCommit,
        author: 'casey@example.com',
        sha: 'e5aef07c91b24d0a8f3310c0e5aef07c91b24d0a',
        shaUrl:
          'https://github.com/acme/platform/commit/e5aef07c91b24d0a8f3310c0e5aef07c91b24d0a',
        createdAt: '2026-10-05T12:00:00Z',
      },
    }}
    changes={changes}
    groups={groups}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
LongCommitMessage.storyName = 'Long commit message'

export const NoPlan = () => (
  <BranchOverview
    hasPlan={false}
    groups={[]}
    rolloutHref="#rollout"
    groupHref={(id) => `#rollout/groups/${id}`}
  />
)
NoPlan.storyName = 'No deployment plan'
