import { useSyncExternalStore } from 'react'

const respondedIds = new Set<string>()
const listeners = new Set<() => void>()

function emit() {
  for (const listener of listeners) {
    listener()
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

let snapshot = 0

export function addRespondedStep(stepId: string) {
  if (!respondedIds.has(stepId)) {
    respondedIds.add(stepId)
    snapshot++
    emit()
  }
}

export function useRespondedApprovals() {
  useSyncExternalStore(subscribe, () => snapshot)

  return {
    hasResponded: (stepId: string) => respondedIds.has(stepId),
  }
}
