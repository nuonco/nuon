export default {
  title: 'Features / Branches / Install groups',
}

import type { TTrackGroup } from '@/components/branches/BranchOverview/RolloutTrack'
import { InstallGroupCards } from './InstallGroupCards'
import { InstallRolloutCard } from './InstallRolloutCard'

const commit = {
  message: 'Add cache component',
  author: 'jane@example.com',
  sha: 'a1b2c3d4e5f60718',
  previousSha: '9f8e7d6c5b4a3928',
  createdAt: '2026-10-08T12:00:00Z',
}

const groups: TTrackGroup[] = [
  {
    id: 'canary',
    name: 'Canary',
    status: 'success',
    maxParallel: 2,
    approval: 'Auto-approves when policies pass',
    match: { kind: 'labels', labels: { env: 'prod', tier: 'canary' } },
    installs: [
      { id: 'alpha', name: 'alpha', status: 'success' },
      { id: 'bravo', name: 'bravo', status: 'success' },
    ],
  },
  {
    id: 'primary',
    name: 'Primary region',
    status: 'in-progress',
    maxParallel: 2,
    approval: 'Manual approval',
    match: {
      kind: 'labels',
      labels: { env: 'prod', region: 'us-east-1', tier: 'primary', ring: '1' },
    },
    installs: [
      { id: 'charlie', name: 'charlie', status: 'success' },
      { id: 'delta', name: 'delta', status: 'error' },
      { id: 'echo', name: 'echo', status: 'in-progress' },
      { id: 'foxtrot', name: 'foxtrot', status: 'approval-awaiting' },
      { id: 'golf', name: 'golf', status: 'pending' },
    ],
  },
  {
    id: 'rest',
    name: 'Remaining',
    status: 'pending',
    maxParallel: 4,
    approval: 'Auto-approves when policies pass',
    match: { kind: 'default' },
    plannedCount: 8,
    installs: [],
  },
]

export const Cards = () => <InstallGroupCards groups={groups} commit={commit} />

export const InstallWaiting = () => (
  <InstallRolloutCard
    install={{ id: 'golf', name: 'golf', status: 'pending' }}
    commit={commit}
  />
)
InstallWaiting.storyName = 'Install has not started'

export const InstallSelectable = () => (
  <InstallRolloutCard
    install={{
      id: 'delta',
      name: 'delta',
      status: 'error',
      workflowId: 'wf-delta',
    }}
    commit={commit}
    onSelect={() => {}}
  />
)
InstallSelectable.storyName = 'Install opens deployment'
