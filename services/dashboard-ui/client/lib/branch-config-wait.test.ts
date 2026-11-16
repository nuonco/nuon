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

  test('stops when the latest of several runs has failed', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      runs: [
        {
          status: {
            status: 'failed',
            status_human_description: 'workflow stopped: missing file',
          },
        },
        { status: { status: 'success' } },
      ],
    })
    expect(state.phase).toBe('failed')
    expect(state.message).toBe('workflow stopped: missing file')
  })

  test('stops when the latest run was cancelled', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      runs: [
        { status: { status: 'cancelled' } },
        { status: { status: 'success' } },
      ],
    })
    expect(state.phase).toBe('failed')
    expect(state.message).toBe(
      'This branch run failed before an app config was created.'
    )
  })

  test('stops when the latest run was not attempted', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      runs: [{ status: { status: 'not-attempted' } }],
    })
    expect(state.phase).toBe('failed')
    expect(state.message).toBe(
      'The latest branch run was skipped, so this branch has no app config. Run the branch again.'
    )
  })

  test('stops when the latest run succeeded but its app config failed to build', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      latestConfigStatus: 'error',
      runs: [{ status: { status: 'success' } }],
    })
    expect(state.phase).toBe('failed')
    expect(state.message).toBe("This branch's app config failed to build.")
  })

  test('keeps waiting while a newer run is in flight after a failed run', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      runs: [
        { status: { status: 'in-progress' } },
        { status: { status: 'failed' } },
      ],
    })
    expect(state.phase).toBe('waiting')
  })

  test('keeps waiting while the latest run succeeded and its config is still syncing', () => {
    const state = branchConfigWaitState({
      hasBranch: true,
      hasActiveConfig: false,
      latestConfigStatus: 'syncing',
      runs: [{ status: { status: 'success' } }],
    })
    expect(state.phase).toBe('waiting')
  })
})
