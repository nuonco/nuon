import { useContext } from 'react'
import { RunnerJobContext } from '@/providers/runner-job-provider'
import type { TRunnerJob } from '@/types'

export function useRunnerJob(): { job: TRunnerJob } {
  const ctx = useContext(RunnerJobContext)
  if (!ctx) {
    throw new Error('useRunnerJob must be used within a RunnerJobProvider')
  }
  return ctx
}
