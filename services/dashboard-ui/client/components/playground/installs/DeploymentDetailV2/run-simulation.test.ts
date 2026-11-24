import { describe, expect, it } from 'bun:test'
import { RUN_SCRIPT } from './run-script'
import {
  advanceRun,
  initRun,
  toWorkflowSteps,
  workflowStatusOf,
  type IRunState,
  type TRunAction,
} from './run-simulation'

type TTestAction =
  | { type: 'tick'; autoApprove?: boolean }
  | { type: 'approve' }
  | { type: 'deny' }

const run = (state: IRunState, action: TTestAction) =>
  advanceRun(RUN_SCRIPT, state, { now: 1_000, ...action } as TRunAction)

const tickUntil = (
  state: IRunState,
  done: (state: IRunState) => boolean,
  autoApprove = false
) => {
  let next = state
  for (let i = 0; i < 500 && !done(next); i++)
    next = run(next, { type: 'tick', autoApprove })
  return next
}

describe('advanceRun', () => {
  it('pauses on the first plan that needs approval', () => {
    const state = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.phase !== 'running'
    )

    expect(state.phase).toBe('awaiting')
    expect(state.steps['run-stack-plan'].status).toBe('approval-awaiting')
    expect(state.steps['run-stack-apply'].status).toBe('pending')
    expect(run(state, { type: 'tick' })).toBe(state)
  })

  it('continues after approval and pauses again on the worker plan', () => {
    const paused = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.phase !== 'running'
    )
    const approved = run(paused, { type: 'approve' })

    expect(approved.steps['run-stack-plan'].status).toBe('success')
    expect(approved.steps['run-stack-apply'].status).toBe('in-progress')

    const second = tickUntil(approved, (s) => s.phase !== 'running')
    expect(second.steps['run-acme-worker-plan'].status).toBe(
      'approval-awaiting'
    )
  })

  it('runs to completion when approvals are automatic', () => {
    const state = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.phase === 'done',
      true
    )

    expect(workflowStatusOf(state.phase)).toBe('success')
    expect(
      Object.values(state.steps).every((step) => step.status === 'success')
    ).toBe(true)
    expect(state.logs['run-acme-web-apply'].length).toBeGreaterThan(0)
  })

  it('stops the run when a plan is denied', () => {
    const paused = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.phase !== 'running'
    )
    const denied = run(paused, { type: 'deny' })

    expect(denied.phase).toBe('denied')
    expect(denied.steps['run-stack-plan'].status).toBe('approval-denied')
    expect(denied.steps['run-acme-web-apply'].status).toBe('not-attempted')
    expect(run(denied, { type: 'tick' })).toBe(denied)
  })
})

describe('toWorkflowSteps', () => {
  it('marks only completed approval steps as responded', () => {
    const paused = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.phase !== 'running'
    )
    const awaiting = toWorkflowSteps(RUN_SCRIPT, paused).find(
      (step) => step.id === 'run-stack-plan'
    )
    expect(awaiting?.approval?.response).toBeUndefined()

    const approved = toWorkflowSteps(
      RUN_SCRIPT,
      run(paused, { type: 'approve' })
    ).find((step) => step.id === 'run-stack-plan')
    expect(approved?.approval?.response).toBeTruthy()
    expect(approved?.finished).toBe(true)
  })
})
