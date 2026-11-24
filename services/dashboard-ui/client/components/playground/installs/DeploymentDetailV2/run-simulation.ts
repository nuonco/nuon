import type { TOTELLog, TWorkflowStep } from '@/types'
import type { IScriptStep, TScriptLevel } from './run-script'

export type TRunPhase = 'running' | 'awaiting' | 'blocked' | 'done' | 'denied'

const RETRY_WAIT_TICKS = 3

export interface IRunAttempt {
  number: number
  status: string
  description?: string
  startedAt?: string
  executionNs?: number
  logs: TOTELLog[]
}

interface IStepRun {
  status: string
  attempt: number
  retryIn?: number
  description?: string
  startedAt?: string
  finishedAt?: string
  executionNs?: number
}

export interface IRunState {
  phase: TRunPhase
  stepIdx: number
  lineIdx: number
  seq: number
  steps: Record<string, IStepRun>
  logs: Record<string, TOTELLog[]>
  attempts: Record<string, IRunAttempt[]>
}

export type TRunAction =
  | { type: 'tick'; now: number; autoApprove?: boolean }
  | { type: 'approve'; now: number }
  | { type: 'deny'; now: number }
  | { type: 'retry'; now: number }
  | { type: 'skip'; now: number }

const SEVERITY: Record<TScriptLevel, { number: number; text: string }> = {
  info: { number: 9, text: 'Info' },
  warn: { number: 13, text: 'Warn' },
  error: { number: 17, text: 'Error' },
}

const iso = (now: number) => new Date(now).toISOString()

export const initRun = (script: IScriptStep[], now: number): IRunState => ({
  phase: 'running',
  stepIdx: 0,
  lineIdx: 0,
  seq: 0,
  steps: Object.fromEntries(
    script.map((step, index) => [
      step.id,
      index === 0
        ? { status: 'in-progress', attempt: 1, startedAt: iso(now) }
        : { status: 'pending', attempt: 1 },
    ])
  ),
  logs: Object.fromEntries(script.map((step) => [step.id, []])),
  attempts: {},
})

const attemptLines = (step: IScriptStep, attempt: number) =>
  attempt <= (step.failAttempts ?? 0)
    ? [...step.lines.slice(0, 2), ...(step.failureLines ?? [])]
    : step.lines

const SKIP_REASON = 'Skipped after repeated Helm failures'

const finishStep = (
  script: IScriptStep[],
  state: IRunState,
  now: number,
  status = 'success'
): IRunState => {
  const current = script[state.stepIdx]
  const run = state.steps[current.id]
  const steps = {
    ...state.steps,
    [current.id]: {
      ...run,
      status,
      description: status === 'user-skipped' ? SKIP_REASON : run.description,
      finishedAt: iso(now),
      executionNs: (now - Date.parse(run.startedAt ?? iso(now))) * 1_000_000,
    },
  }
  const next = script[state.stepIdx + 1]

  if (!next) return { ...state, steps, phase: 'done' }

  return {
    ...state,
    steps: {
      ...steps,
      [next.id]: { status: 'in-progress', attempt: 1, startedAt: iso(now) },
    },
    stepIdx: state.stepIdx + 1,
    lineIdx: 0,
    phase: 'running',
  }
}

const denyStep = (
  script: IScriptStep[],
  state: IRunState,
  now: number
): IRunState => {
  const steps = Object.fromEntries(
    script.map((step, index) => [
      step.id,
      index === state.stepIdx
        ? {
            ...state.steps[step.id],
            status: 'approval-denied',
            finishedAt: iso(now),
          }
        : index > state.stepIdx
          ? { status: 'not-attempted', attempt: 1 }
          : state.steps[step.id],
    ])
  )
  return { ...state, steps, phase: 'denied' }
}

const appendLines = (
  script: IScriptStep[],
  state: IRunState,
  now: number
): IRunState => {
  const step = script[state.stepIdx]
  const lines = attemptLines(step, state.steps[step.id].attempt)
  const batch = lines.slice(state.lineIdx, state.lineIdx + (step.batch ?? 1))
  const logs = batch.map((line, index) => {
    const severity = SEVERITY[line.level]
    return {
      id: `${step.id}-log-${state.seq + index}`,
      timestamp: iso(now + index),
      body: line.text,
      severity_number: severity.number,
      severity_text: severity.text,
      service_name: 'runner',
      scope_name: 'oteljob',
    } as TOTELLog
  })

  return {
    ...state,
    seq: state.seq + logs.length,
    lineIdx: state.lineIdx + logs.length,
    logs: { ...state.logs, [step.id]: [...state.logs[step.id], ...logs] },
  }
}

const withRun = (
  state: IRunState,
  step: IScriptStep,
  run: IStepRun
): IRunState => ({ ...state, steps: { ...state.steps, [step.id]: run } })

const startRetry = (
  state: IRunState,
  step: IScriptStep,
  now: number
): IRunState => {
  const run = state.steps[step.id]
  const prior: IRunAttempt = {
    number: run.attempt,
    status: 'error',
    description: run.description,
    startedAt: run.startedAt,
    executionNs: run.executionNs,
    logs: state.logs[step.id],
  }

  return {
    ...withRun(state, step, {
      status: 'in-progress',
      attempt: run.attempt + 1,
      startedAt: iso(now),
    }),
    lineIdx: 0,
    logs: { ...state.logs, [step.id]: [] },
    attempts: {
      ...state.attempts,
      [step.id]: [...(state.attempts[step.id] ?? []), prior],
    },
  }
}

export const advanceRun = (
  script: IScriptStep[],
  state: IRunState,
  action: TRunAction
): IRunState => {
  if (state.phase === 'done' || state.phase === 'denied') return state

  if (state.phase === 'blocked') {
    if (action.type === 'skip')
      return finishStep(script, state, action.now, 'user-skipped')
    if (action.type === 'retry')
      return {
        ...startRetry(state, script[state.stepIdx], action.now),
        phase: 'running',
      }
    return state
  }

  if (state.phase === 'awaiting') {
    if (action.type === 'approve') return finishStep(script, state, action.now)
    if (action.type === 'deny') return denyStep(script, state, action.now)
    if (action.type === 'tick' && action.autoApprove)
      return finishStep(script, state, action.now)
    return state
  }

  if (action.type !== 'tick') return state

  const step = script[state.stepIdx]
  const run = state.steps[step.id]

  if (run.retryIn !== undefined)
    return run.retryIn > 1
      ? withRun(state, step, { ...run, retryIn: run.retryIn - 1 })
      : startRetry(state, step, action.now)

  const lines = attemptLines(step, run.attempt)
  if (state.lineIdx < lines.length)
    return appendLines(script, state, action.now)

  if (step.awaitAction && run.attempt === step.failAttempts)
    return {
      ...withRun(state, step, {
        ...run,
        status: 'failed-pending-retry',
        description: step.failureDescription,
        finishedAt: iso(action.now),
        executionNs:
          (action.now - Date.parse(run.startedAt ?? iso(action.now))) *
          1_000_000,
      }),
      phase: 'blocked',
    }

  if (run.attempt <= (step.failAttempts ?? 0))
    return withRun(state, step, {
      ...run,
      status: 'failed-pending-retry',
      description: step.failureDescription,
      finishedAt: iso(action.now),
      executionNs:
        (action.now - Date.parse(run.startedAt ?? iso(action.now))) * 1_000_000,
      retryIn: RETRY_WAIT_TICKS,
    })

  if (step.prompt && !action.autoApprove)
    return {
      ...state,
      phase: 'awaiting',
      steps: {
        ...state.steps,
        [step.id]: { ...state.steps[step.id], status: 'approval-awaiting' },
      },
    }

  return finishStep(script, state, action.now)
}

export const workflowStatusOf = (phase: TRunPhase): string =>
  ({
    running: 'in-progress',
    awaiting: 'approval-awaiting',
    blocked: 'failed-pending-retry',
    done: 'success',
    denied: 'error',
  })[phase]

export const toWorkflowSteps = (
  script: IScriptStep[],
  state: IRunState
): TWorkflowStep[] =>
  script.map((step, index) => {
    const run = state.steps[step.id]
    const finished = run.status === 'success' || run.status === 'user-skipped'
    const blocked = state.phase === 'blocked' && state.stepIdx === index
    const reviewed = finished || run.status === 'approval-awaiting'
    const policyMeta =
      step.policy && (reviewed || run.status === 'approval-denied')
        ? {
            deny_violations: [],
            warn_violations: (step.policy.warnings ?? []).map(
              (message, idx) => ({
                policy_id: `${step.id}-warn-${idx}`,
                message,
                severity: 'warn',
              })
            ),
            skip_findings: (step.policy.skipped ?? []).map((message, idx) => ({
              policy_id: `${step.id}-skip-${idx}`,
              message,
            })),
            passed_policy_ids: [],
          }
        : undefined
    const retryMeta =
      run.attempt > 1
        ? { is_retry: true, retry_idx: run.attempt - 1, retry_type: 'auto' }
        : undefined
    const statusMeta =
      policyMeta || retryMeta ? { ...policyMeta, ...retryMeta } : undefined

    return {
      id: step.id,
      name: step.name,
      group_idx: step.componentName
        ? script.findIndex(
            (other) => other.componentName === step.componentName
          )
        : index,
      step_target_type: step.targetType,
      execution_type: step.approval ? 'approval' : 'system',
      finished,
      retryable: blocked,
      skippable: blocked,
      created_by:
        run.status === 'user-skipped'
          ? { email: 'ada@example.com' }
          : undefined,
      started_at: run.startedAt,
      execution_time: run.executionNs,
      status: {
        status: run.status,
        status_human_description: run.description,
        metadata: statusMeta,
      },
      metadata: step.componentName
        ? { component_name: step.componentName }
        : {},
      approval:
        step.approval &&
        step.changes &&
        (reviewed || run.status === 'approval-denied')
          ? {
              id: `approval-${step.id}`,
              type: step.approval,
              changes_state: 'ok',
              changes_create: step.changes.create,
              changes_update: step.changes.update,
              changes_delete: step.changes.delete,
              changes_replace: 0,
              changes_noop: step.changes.noop,
              response: finished ? { id: `response-${step.id}` } : undefined,
            }
          : undefined,
    } as TWorkflowStep
  })
