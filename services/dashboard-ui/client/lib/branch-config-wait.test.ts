import { describe, expect, test } from 'bun:test'
import { branchConfigWaitState } from './branch-config-wait'

describe('branchConfigWaitState', () => {
  test('skips the wait when no branch is selected', () => {
    expect(
      branchConfigWaitState({ hasBranch: false, hasActiveConfig: false }).phase
    ).toBe('skip')
  })

  test('is ready once the selected branch has an active config', () => {
    expect(
      branchConfigWaitState({ hasBranch: true, hasActiveConfig: true }).phase
    ).toBe('ready')
  })

  test('keeps waiting while the only run is still in flight', () => {
    expect(
      branchConfigWaitState({
        hasBranch: true,
        hasActiveConfig: false,
        runs: [{ status: { status: 'running' } }],
      }).phase
    ).toBe('waiting')
  })

  test('stops when the only branch run has failed', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      runs: [
        {
          status: {
            status: 'error',
            status_human_description: 'config failed to build',
          },
        },
      ],
    })
    expect(state.phase).toBe('failed')
    expect(state.message).toBe('config failed to build')
  })
})
