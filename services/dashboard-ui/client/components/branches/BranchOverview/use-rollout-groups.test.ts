import { describe, expect, test } from 'bun:test'
import type { TInstallGroupRun, TInstallWorkflowStep } from '@/types'
import type { TTrackGroup } from './RolloutTrack'
import type { TAppBranchRun, TInstall } from '@/types'
import {
  historicalRunGroups,
  historyPreviewMode,
  mergeGroupRuns,
  previewAffectedInstalls,
  rolloutGroupHrefForWorkflow,
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

describe('mergeGroupRuns install detail', () => {
  test('describes customer waits, queues, and superseded runs', () => {
    const [group] = mergeGroupRuns(
      [],
      [
        {
          id: 'group-run-1',
          install_group_id: 'primary',
          installs: [
            {
              install_id: 'ins_wait',
              status: 'pending_customer',
              release_reason: 'stack_pending_customer',
            },
            {
              install_id: 'ins_runner',
              status: 'pending_customer',
              release_reason: 'runner_offline',
            },
            {
              install_id: 'ins_queue',
              status: 'queued',
              waiting_on_run_id: 'run_earlier',
            },
            { install_id: 'ins_old', status: 'superseded' },
          ],
        } as TInstallGroupRun,
      ],
      [],
      {}
    )
    expect(group.installs.map((install) => install.detail)).toEqual([
      'Waiting on customer',
      'Runner offline',
      'Queued behind run run_earlier',
      'Superseded by a later run',
    ])
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

describe('rolloutGroupHrefForWorkflow', () => {
  test('opens a group under the branch rollout', () => {
    expect(
      rolloutGroupHrefForWorkflow(
        '/org-1/apps/app-1/branches/branch-1',
        'canary'
      )
    ).toBe('/org-1/apps/app-1/branches/branch-1/rollout/groups/canary')
  })

  test('keeps a previous run on that run group page', () => {
    expect(
      rolloutGroupHrefForWorkflow(
        '/org-1/apps/app-1/branches/branch-1',
        'canary',
        'wf-1'
      )
    ).toBe(
      '/org-1/apps/app-1/branches/branch-1/runs/wf-1/rollout/groups/canary'
    )
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

describe('historyPreviewMode', () => {
  test('ignores a preview unless the run is opened from history', () => {
    expect(
      historyPreviewMode(
        { preview: { mode: 'build-only' } } as TAppBranchRun,
        false
      )
    ).toBeUndefined()
  })

  test('keeps build and validate, plan, and apply', () => {
    expect(
      historyPreviewMode(
        { preview: { mode: 'build-only' } } as TAppBranchRun,
        true
      )
    ).toBe('build-only')
    expect(
      historyPreviewMode({ preview: { mode: 'apply' } } as TAppBranchRun, true)
    ).toBe('apply')
  })
})

describe('previewAffectedInstalls', () => {
  const installs = [
    { id: 'ins_a', name: 'alpha', labels: { tier: 'canary' } },
    { id: 'ins_b', name: 'bravo', labels: { tier: 'prod' } },
  ] as TInstall[]

  test('returns the named install', () => {
    expect(
      previewAffectedInstalls(installs, {
        preview: { install_id: 'ins_b' },
      } as TAppBranchRun).map((install) => install.id)
    ).toEqual(['ins_b'])
  })

  test('returns installs matching the preview selector', () => {
    expect(
      previewAffectedInstalls(installs, {
        preview: {
          resolved_preview_config: {
            label_selector: { match_labels: { tier: 'canary' } },
          },
        },
      } as TAppBranchRun).map((install) => install.id)
    ).toEqual(['ins_a'])
  })
})
