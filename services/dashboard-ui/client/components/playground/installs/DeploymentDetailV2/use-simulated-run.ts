import { useCallback, useEffect, useMemo, useState } from 'react'
import type { TOTELLog, TWorkflowStep } from '@/types'
import { RUN_SCRIPT } from './run-script'
import {
  advanceRun,
  initRun,
  type IRunAttempt,
  toWorkflowSteps,
  workflowStatusOf,
  type TRunAction,
} from './run-simulation'

export interface IUseSimulatedRun {
  tickMs?: number
  autoApprove?: boolean
}

export interface ISimulatedRun {
  steps: TWorkflowStep[]
  logs: Record<string, TOTELLog[]>
  attempts: Record<string, IRunAttempt[]>
  status: string
  isFinished: boolean
  approve: () => void
  deny: () => void
  retry: () => void
  skip: () => void
  restart: () => void
}

export const useSimulatedRun = ({
  tickMs = 800,
  autoApprove = false,
}: IUseSimulatedRun = {}): ISimulatedRun => {
  const [state, setState] = useState(() => initRun(RUN_SCRIPT, Date.now()))
  const isFinished = state.phase === 'done' || state.phase === 'denied'

  const dispatch = useCallback(
    (action: TRunAction) =>
      setState((current) => advanceRun(RUN_SCRIPT, current, action)),
    []
  )

  useEffect(() => {
    if (isFinished) return
    const id = setInterval(
      () => dispatch({ type: 'tick', now: Date.now(), autoApprove }),
      tickMs
    )
    return () => clearInterval(id)
  }, [autoApprove, dispatch, isFinished, tickMs])

  const steps = useMemo(() => toWorkflowSteps(RUN_SCRIPT, state), [state])

  return {
    steps,
    logs: state.logs,
    attempts: state.attempts,
    status: workflowStatusOf(state.phase),
    isFinished,
    approve: () => dispatch({ type: 'approve', now: Date.now() }),
    deny: () => dispatch({ type: 'deny', now: Date.now() }),
    retry: () => dispatch({ type: 'retry', now: Date.now() }),
    skip: () => dispatch({ type: 'skip', now: Date.now() }),
    restart: () => setState(initRun(RUN_SCRIPT, Date.now())),
  }
}
