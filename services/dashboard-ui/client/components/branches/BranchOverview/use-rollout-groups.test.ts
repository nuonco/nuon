import { describe, expect, test } from 'bun:test'
import type { TInstallGroupRun } from '@/types'
import type { TTrackGroup } from './RolloutTrack'
import { mergeGroupRuns } from './use-rollout-groups'

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
