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
  | { type: 'retry' }
  | { type: 'skip' }

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

  const blocked = () =>
    tickUntil(initRun(RUN_SCRIPT, 0), (s) => s.phase === 'blocked', true)

  it('blocks on a step that needs a manual retry or skip', () => {
    const state = blocked()

    expect(workflowStatusOf(state.phase)).toBe('failed-pending-retry')
    expect(state.steps['run-acme-web-apply'].status).toBe(
      'failed-pending-retry'
    )
    expect(state.attempts['run-acme-web-apply']).toHaveLength(2)
    expect(run(state, { type: 'tick', autoApprove: true })).toBe(state)

    const steps = toWorkflowSteps(RUN_SCRIPT, state)
    const step = steps.find((item) => item.id === 'run-acme-web-apply')
    expect(step?.skippable).toBe(true)
    expect(step?.retryable).toBe(true)
  })

  it('continues past a manually skipped step', () => {
    const skipped = run(blocked(), { type: 'skip' })
    const step = skipped.steps['run-acme-web-apply']

    expect(step.status).toBe('user-skipped')
    expect(skipped.steps['run-verify'].status).toBe('in-progress')

    const done = tickUntil(skipped, (s) => s.phase === 'done', true)
    expect(workflowStatusOf(done.phase)).toBe('success')
    expect(done.steps['run-verify'].status).toBe('success')
  })

  it('succeeds on a manual retry', () => {
    const retried = run(blocked(), { type: 'retry' })
    expect(retried.steps['run-acme-web-apply'].attempt).toBe(4)

    const done = tickUntil(retried, (s) => s.phase === 'done', true)
    expect(done.steps['run-acme-web-apply'].status).toBe('success')
    expect(done.attempts['run-acme-web-apply']).toHaveLength(3)
  })

  it('retries a failing step until it succeeds', () => {
    const failing = tickUntil(
      initRun(RUN_SCRIPT, 0),
      (s) => s.steps['run-acme-worker-apply'].status === 'failed-pending-retry',
      true
    )
    expect(failing.steps['run-acme-worker-apply'].attempt).toBe(1)

    const done = tickUntil(
      failing,
      (s) => s.steps['run-acme-worker-apply'].status === 'success',
      true
    )
    const step = done.steps['run-acme-worker-apply']
    expect(step.attempt).toBe(3)
    expect(done.attempts['run-acme-worker-apply'].map((a) => a.status)).toEqual(
      ['error', 'error']
    )
    expect(done.logs['run-acme-worker-apply'].at(-1)?.body).toBe(
      'Release upgraded'
    )
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
