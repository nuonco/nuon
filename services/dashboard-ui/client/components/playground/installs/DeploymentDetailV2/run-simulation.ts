import type { TOTELLog, TWorkflowStep } from '@/types'
import type { IScriptStep, TScriptLevel } from './run-script'

export type TRunPhase = 'running' | 'awaiting' | 'done' | 'denied'

interface IStepRun {
  status: string
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
}

export type TRunAction =
  | { type: 'tick'; now: number; autoApprove?: boolean }
  | { type: 'approve'; now: number }
  | { type: 'deny'; now: number }

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
        ? { status: 'in-progress', startedAt: iso(now) }
        : { status: 'pending' },
    ])
  ),
  logs: Object.fromEntries(script.map((step) => [step.id, []])),
})

const finishStep = (
  script: IScriptStep[],
  state: IRunState,
  now: number
): IRunState => {
  const current = script[state.stepIdx]
  const run = state.steps[current.id]
  const steps = {
    ...state.steps,
    [current.id]: {
      ...run,
      status: 'success',
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
      [next.id]: { status: 'in-progress', startedAt: iso(now) },
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
          ? { status: 'not-attempted' }
          : state.steps[step.id],
    ])
  )
  return { ...state, steps, phase: 'denied' }
}

const appendLine = (
  script: IScriptStep[],
  state: IRunState,
  now: number
): IRunState => {
  const step = script[state.stepIdx]
  const line = step.lines[state.lineIdx]
  const severity = SEVERITY[line.level]
  const log = {
    id: `${step.id}-log-${state.seq}`,
    timestamp: iso(now),
    body: line.text,
    severity_number: severity.number,
    severity_text: severity.text,
    service_name: 'runner',
    scope_name: 'oteljob',
  } as TOTELLog

  return {
    ...state,
    seq: state.seq + 1,
    lineIdx: state.lineIdx + 1,
    logs: { ...state.logs, [step.id]: [...state.logs[step.id], log] },
  }
}

export const advanceRun = (
  script: IScriptStep[],
  state: IRunState,
  action: TRunAction
): IRunState => {
  if (state.phase === 'done' || state.phase === 'denied') return state

  if (state.phase === 'awaiting') {
    if (action.type === 'approve') return finishStep(script, state, action.now)
    if (action.type === 'deny') return denyStep(script, state, action.now)
    if (action.type === 'tick' && action.autoApprove)
      return finishStep(script, state, action.now)
    return state
  }

  if (action.type !== 'tick') return state

  const step = script[state.stepIdx]
  if (state.lineIdx < step.lines.length)
    return appendLine(script, state, action.now)

  if (step.prompt)
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
    done: 'success',
    denied: 'error',
  })[phase]

export const toWorkflowSteps = (
  script: IScriptStep[],
  state: IRunState
): TWorkflowStep[] =>
  script.map((step, index) => {
    const run = state.steps[step.id]
    const finished = run.status === 'success'
    const reviewed = finished || run.status === 'approval-awaiting'

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
      started_at: run.startedAt,
      execution_time: run.executionNs,
      status: { status: run.status },
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
