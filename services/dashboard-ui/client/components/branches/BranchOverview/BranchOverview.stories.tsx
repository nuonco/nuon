export default {
  title: 'Features / Branches / Branch overview',
  fullBleed: true,
}

import { Text } from '@/components/common/Text'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { BranchOverview, type TOverviewRollout } from './BranchOverview'
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
  <div className="flex flex-col gap-3">
    <SectionHeader title="What's changed" />
    <Text variant="subtext" theme="neutral">
      cache added, api image tag 1.4.2
    </Text>
  </div>
)

export const RollingOut = () => (
  <BranchOverview
    hasPlan
    rollout={rollout}
    changes={changes}
    groups={groups}
    rolloutHref="#rollout"
    onSelectGroup={() => {}}
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
    onSelectGroup={() => {}}
  />
)
TagPush.storyName = 'Tag push'

export const NoPlan = () => (
  <BranchOverview
    hasPlan={false}
    groups={[]}
    rolloutHref="#rollout"
    onSelectGroup={() => {}}
  />
)
NoPlan.storyName = 'No deployment plan'
