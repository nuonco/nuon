import { describe, expect, test } from 'bun:test'
import type { TInstallGroupRun, TInstallWorkflowStep } from '@/types'
import type { TTrackGroup } from './RolloutTrack'
import {
  historicalRunGroups,
  mergeGroupRuns,
  rolloutHrefForWorkflow,
} from './use-rollout-groups'

const planned = (id: string, name: string): TTrackGroup => ({
  id,
  name,
  status: 'pending',
  installs: [
    { id: `${id}-install`, name: `${name} install`, status: 'pending' },
  ],
})

describe('mergeGroupRuns', () => {
  test('keeps groups that have not started when another group failed', () => {
    const groups = mergeGroupRuns(
      [
        planned('canary', 'Canary'),
        planned('primary', 'Primary region'),
        planned('rest', 'Remaining'),
      ],
      [
        {
          id: 'run-1',
          install_group_id: 'primary',
          install_group_name: 'Primary region',
          status: { status: 'error' },
          installs: [{ install_id: 'ins_1', status: 'error' }],
        } as TInstallGroupRun,
      ],
      [],
      {}
    )
    expect(groups.map((group) => [group.id, group.status])).toEqual([
      ['canary', 'pending'],
      ['primary', 'error'],
      ['rest', 'pending'],
    ])
    expect(groups[2].installs).toHaveLength(1)
  })
})

describe('rolloutHrefForWorkflow', () => {
  test('opens the branch rollout for the current run', () => {
    expect(rolloutHrefForWorkflow('/org-1/apps/app-1/branches/branch-1')).toBe(
      '/org-1/apps/app-1/branches/branch-1/rollout'
    )
  })

  test('keeps a previous run on that run rollout', () => {
    expect(
      rolloutHrefForWorkflow('/org-1/apps/app-1/branches/branch-1', 'wf-1')
    ).toBe('/org-1/apps/app-1/branches/branch-1/runs/wf-1/rollout')
  })
})

describe('historicalRunGroups', () => {
  test('drops install groups added to the plan after this run', () => {
    const groups = historicalRunGroups(
      [planned('canary', 'Canary'), planned('later', 'Later')],
      [
        {
          name: 'plan install group: Canary',
          status: { status: 'pending' },
        } as TInstallWorkflowStep,
      ]
    )
    expect(groups.map((group) => group.id)).toEqual(['canary'])
  })

  test('keeps groups this run already finished', () => {
    const groups = historicalRunGroups(
      [
        { ...planned('canary', 'Canary'), status: 'success' },
        planned('later', 'Later'),
      ],
      []
    )
    expect(groups.map((group) => group.id)).toEqual(['canary'])
  })
})
