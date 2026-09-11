import { describe, expect, test } from 'bun:test'
import type { TAppBranchInstallGroup, TInstallGroupRun } from '@/types'
import { resolveDeploymentPlanStages } from './deployment-plan'

const group = (
  fields: TAppBranchInstallGroup
): TAppBranchInstallGroup => fields

const run = (fields: TInstallGroupRun): TInstallGroupRun => fields

describe('resolveDeploymentPlanStages', () => {
  test('returns groups ordered by order', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({ id: 'g-late', name: 'canary', order: 2 }),
        group({ id: 'g-first', name: 'core', order: 0 }),
        group({ id: 'g-mid', name: 'edge', order: 1 }),
      ],
      installs: [],
    })

    expect(stages.map((stage) => stage.id)).toEqual([
      'g-first',
      'g-mid',
      'g-late',
    ])
    expect(stages.map((stage) => stage.stage)).toEqual([1, 2, 3])
  })

  test('assigns an install matched by two groups to the earlier group only', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-stage',
          name: 'stage',
          order: 0,
          label_selector: { match_labels: { env: 'stage' } },
        }),
        group({
          id: 'g-any-env',
          name: 'rest',
          order: 1,
          label_selector: { match_labels: { env: '*' } },
        }),
      ],
      installs: [
        { id: 'ins-stage', name: 'payments-stage', labels: { env: 'stage' } },
        { id: 'ins-prod', name: 'payments-prod', labels: { env: 'prod' } },
      ],
    })

    expect(stages[0]?.installs.map((install) => install.id)).toEqual([
      'ins-stage',
    ])
    expect(stages[1]?.installs.map((install) => install.id)).toEqual([
      'ins-prod',
    ])
  })

  test('all_installs claims what no earlier group claimed', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-ids',
          name: 'named',
          order: 0,
          install_ids: ['ins-a'],
        }),
        group({
          id: 'g-all',
          name: 'everyone else',
          order: 1,
          all_installs: true,
        }),
      ],
      installs: [
        { id: 'ins-a', name: 'alpha' },
        { id: 'ins-b', name: 'bravo' },
        { id: 'ins-c', name: 'charlie' },
      ],
    })

    expect(stages[0]?.membership).toBe('install_ids')
    expect(stages[0]?.installs.map((install) => install.id)).toEqual(['ins-a'])
    expect(stages[1]?.membership).toBe('all_installs')
    expect(stages[1]?.installs.map((install) => install.id)).toEqual([
      'ins-b',
      'ins-c',
    ])
  })

  test('resolves install_ids and label_selector', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-ids',
          name: 'fixed',
          order: 0,
          install_ids: ['ins-b', 'ins-a'],
        }),
        group({
          id: 'g-sel',
          name: 'labeled',
          order: 1,
          label_selector: {
            match_labels: { tier: '*' },
            not_match_labels: { env: 'stage' },
          },
        }),
      ],
      installs: [
        { id: 'ins-a', name: 'alpha' },
        { id: 'ins-b', name: 'bravo' },
        {
          id: 'ins-prod',
          name: 'prod',
          labels: { tier: 'a', env: 'prod' },
        },
        {
          id: 'ins-stage',
          name: 'stage',
          labels: { tier: 'a', env: 'stage' },
        },
      ],
    })

    expect(stages[0]?.installs.map((install) => install.id)).toEqual([
      'ins-b',
      'ins-a',
    ])
    expect(stages[1]?.membership).toBe('label_selector')
    expect(stages[1]?.selector).toEqual({
      match_labels: { tier: '*' },
      not_match_labels: { env: 'stage' },
    })
    expect(stages[1]?.installs.map((install) => install.id)).toEqual(['ins-prod'])
  })

  test('omits an install that matches no group', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-prod',
          name: 'prod',
          order: 0,
          label_selector: { match_labels: { env: 'prod' } },
        }),
      ],
      installs: [
        { id: 'ins-prod', labels: { env: 'prod' } },
        { id: 'ins-drift', labels: { env: 'dev' } },
      ],
    })

    expect(stages.flatMap((stage) => stage.installs.map((i) => i.id))).toEqual([
      'ins-prod',
    ])
  })

  test('still produces a stage for an empty group', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-empty',
          name: 'empty',
          order: 0,
          label_selector: { match_labels: { env: 'prod' } },
        }),
      ],
      installs: [{ id: 'ins-stage', labels: { env: 'stage' } }],
    })

    expect(stages).toHaveLength(1)
    expect(stages[0]?.installs).toEqual([])
    expect(stages[0]?.totalInstalls).toBe(0)
  })

  test('lays group run status and counts on the matching group', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-core',
          name: 'core',
          order: 0,
          install_ids: ['ins-a', 'ins-b'],
        }),
        group({
          id: 'g-edge',
          name: 'edge',
          order: 1,
          install_ids: ['ins-c'],
        }),
      ],
      installs: [
        { id: 'ins-a' },
        { id: 'ins-b' },
        { id: 'ins-c' },
        { id: 'ins-d' },
      ],
      groupRuns: [
        run({
          install_group_id: 'g-core',
          total_installs: 12,
          completed_installs: 10,
          failed_installs: 2,
          status: { status: 'error' },
        }),
      ],
    })

    expect(stages[0]?.status).toBe('error')
    expect(stages[0]?.totalInstalls).toBe(12)
    expect(stages[0]?.completedInstalls).toBe(10)
    expect(stages[0]?.failedInstalls).toBe(2)
    expect(stages[0]?.installs).toHaveLength(2)
    expect(stages[1]?.status).toBeUndefined()
    expect(stages[1]?.totalInstalls).toBe(1)
    expect(stages[1]?.completedInstalls).toBeUndefined()
  })

  test('ignores a group run for a group no longer in the config', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-now',
          name: 'now',
          order: 0,
          install_ids: ['ins-a'],
        }),
      ],
      installs: [{ id: 'ins-a' }],
      groupRuns: [
        run({
          install_group_id: 'g-gone',
          total_installs: 99,
          status: { status: 'success' },
        }),
      ],
    })

    expect(stages).toHaveLength(1)
    expect(stages[0]?.status).toBeUndefined()
    expect(stages[0]?.totalInstalls).toBe(1)
  })

  test('carries configured state and no status when there are no runs', () => {
    const stages = resolveDeploymentPlanStages({
      groups: [
        group({
          id: 'g-core',
          name: 'core',
          order: 0,
          max_parallel: 3,
          auto_approve_on_policies_passing: true,
          install_ids: ['ins-a'],
        }),
      ],
      installs: [{ id: 'ins-a', name: 'alpha' }],
    })

    expect(stages[0]?.status).toBeUndefined()
    expect(stages[0]?.completedInstalls).toBeUndefined()
    expect(stages[0]?.failedInstalls).toBeUndefined()
    expect(stages[0]?.totalInstalls).toBe(1)
    expect(stages[0]?.maxParallel).toBe(3)
    expect(stages[0]?.autoApproveOnPoliciesPassing).toBe(true)
  })
})
