import type { TWorkflowStep } from '@/types'

export type TWaitState = 'done' | 'active' | 'pending' | 'error'

export interface IWaitItem {
  label: string
  state: TWaitState
}

export interface IWaitItemDef {
  label: string
  // Workflow step name this item tracks. Items without one are marked by the caller.
  step?: string
  done?: boolean
}

const DONE = new Set(['success', 'skipped', 'user-skipped', 'auto-skipped', 'noop'])
const ACTIVE = new Set(['in-progress', 'running', 'planning', 'applying', 'approval-awaiting'])
const FAILED = new Set(['error', 'failed', 'cancelled', 'outdated'])

const stepState = (step?: TWorkflowStep): TWaitState => {
  const status = step?.status?.status as string | undefined
  if (!status) return 'pending'
  if (DONE.has(status)) return 'done'
  if (FAILED.has(status)) return 'error'
  if (ACTIVE.has(status)) return 'active'
  return 'pending'
}

// Anything before the furthest item that has started counts as done, and the
// first unfinished item shows as active, so the list keeps moving through the
// gaps between workflow steps.
export function waitItems(defs: IWaitItemDef[], steps: TWorkflowStep[] = []): IWaitItem[] {
  const states = defs.map((def): TWaitState => {
    if (def.done) return 'done'
    if (!def.step) return 'pending'
    return stepState(steps.find((step) => step.name === def.step))
  })

  const furthest = states.reduce((last, state, i) => (state === 'pending' ? last : i), -1)
  for (let i = 0; i < furthest; i++) {
    if (states[i] === 'pending' || states[i] === 'active') states[i] = 'done'
  }
  if (!states.includes('error') && !states.includes('active')) {
    const next = states.indexOf('pending')
    if (next !== -1) states[next] = 'active'
  }

  return defs.map((def, i) => ({ label: def.label, state: states[i] }))
}

export const formatElapsed = (ms: number) => {
  const seconds = Math.max(0, Math.floor(ms / 1000))
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`
}
