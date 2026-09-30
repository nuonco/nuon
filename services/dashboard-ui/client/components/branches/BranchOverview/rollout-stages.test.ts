import { describe, expect, test } from 'bun:test'
import type {
  TAppBranchInstallGroup,
  TInstall,
  TInstallWorkflowStep,
} from '@/types'
import {
  buildRolloutStages,
  deployStepForGroup,
  planStepForGroup,
} from './rollout-stages'

const step = (
  over: Partial<TInstallWorkflowStep> & { name: string }
): TInstallWorkflowStep =>
  ({
    id: over.name,
    status: { status: 'pending' },
    ...over,
  }) as TInstallWorkflowStep

const group = (name: string, order: number): TAppBranchInstallGroup => ({
  id: name,
  name,
  order,
})

describe('buildRolloutStages', () => {
  test('shows commit, build, and groups as pending when nothing has run', () => {
    const stages = buildRolloutStages({
      groups: [group('Canary', 1), group('Primary region', 2)],
      steps: [],
      installsByGroup: [
        [{ id: 'ins_alpha', name: 'alpha' } as TInstall],
        [
          { id: 'ins_delta', name: 'delta' } as TInstall,
          { id: 'ins_charlie', name: 'charlie' } as TInstall,
        ],
      ],
    })

    expect(
      stages.map((stage) => [stage.name, stage.status, stage.detail])
    ).toEqual([
      ['Commit', 'pending', undefined],
      ['Build', 'pending', undefined],
      ['Canary', 'pending', '1 install'],
      ['Primary region', 'pending', '2 installs'],
    ])
  })

  test('maps commit, build, and a deploy group from workflow steps', () => {
    const stages = buildRolloutStages({
      groups: [group('Canary', 1)],
      sha: 'a1b2c3d4e5',
      steps: [
        step({
          name: 'record commit',
          status: { status: 'success' },
          finished: true,
          execution_time: 4_000_000_000,
        }),
        step({
          name: 'building components and sandbox',
          status: { status: 'in-progress' },
        }),
        step({
          name: 'plan install group: Canary',
          status: { status: 'success' },
          finished: true,
        }),
        step({
          id: 'deploy-canary',
          name: 'deploy install group: Canary',
          status: {
            status: 'in-progress',
            metadata: {
              installs: [
                { install_id: 'ins_alpha', status: 'success' },
                { install_id: 'ins_bravo', status: 'error' },
              ],
            },
          },
        }),
      ],
      installsByGroup: [
        [
          {
            id: 'ins_alpha',
            name: 'alpha',
            aws_account: { region: 'us-east-1' },
          } as TInstall,
          { id: 'ins_bravo', name: 'bravo' } as TInstall,
        ],
      ],
    })

    expect(stages[0]).toMatchObject({
      kind: 'commit',
      status: 'success',
      detail: 'a1b2c3d',
      durationNs: 4_000_000_000,
    })
    expect(stages[1]).toMatchObject({ kind: 'build', status: 'in-progress' })
    expect(stages[2]).toMatchObject({
      kind: 'group',
      name: 'Canary',
      status: 'error',
      stepId: 'deploy-canary',
    })
    expect(stages[2].installs).toEqual([
      {
        id: 'ins_alpha',
        name: 'alpha',
        status: 'success',
        region: 'us-east-1',
      },
      { id: 'ins_bravo', name: 'bravo', status: 'error', region: undefined },
    ])
  })

  test('keeps install counts with the group after ordering', () => {
    const stages = buildRolloutStages({
      groups: [group('Primary region', 2), group('Canary', 1)],
      steps: [],
      installsByGroup: [
        [
          { id: 'ins_delta', name: 'delta' } as TInstall,
          { id: 'ins_charlie', name: 'charlie' } as TInstall,
        ],
        [{ id: 'ins_alpha', name: 'alpha' } as TInstall],
      ],
    })

    expect(stages.map((stage) => [stage.name, stage.detail])).toEqual([
      ['Commit', undefined],
      ['Build', undefined],
      ['Canary', '1 install'],
      ['Primary region', '2 installs'],
    ])
  })

  test('treats a finished plan with a queued deploy as in progress', () => {
    const stages = buildRolloutStages({
      groups: [group('Remaining', 1)],
      steps: [
        step({
          name: 'plan install group: remaining',
          status: { status: 'success' },
          finished: true,
        }),
        step({
          name: 'deploy install group: Remaining',
          status: { status: 'pending' },
        }),
      ],
      installsByGroup: [[]],
    })

    expect(stages[2].status).toBe('in-progress')
  })
})

describe('planStepForGroup', () => {
  test('returns only the plan step for that group', () => {
    const steps = [
      step({ id: 'plan-1', name: 'plan install group: group-1' }),
      step({ id: 'deploy-1', name: 'deploy install group: group-1' }),
      step({ id: 'plan-2', name: 'plan install group: group-2' }),
    ]
    expect(planStepForGroup(steps, 'group-1')?.id).toBe('plan-1')
    expect(planStepForGroup(steps, 'Group-2')?.id).toBe('plan-2')
    expect(planStepForGroup(steps, 'preview')).toBeUndefined()
    expect(deployStepForGroup(steps, 'group-1')?.id).toBe('deploy-1')
    expect(deployStepForGroup(steps, 'group-2')).toBeUndefined()
  })
})
