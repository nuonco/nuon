import { useContext } from 'react'
import { RunnerContext } from '@/providers/runner-provider'
import type { TRunner } from '@/types'

export function useRunner(): { runner: TRunner } {
  const ctx = useContext(RunnerContext)
  if (!ctx) {
    throw new Error('useRunner must be used within a RunnerProvider')
  }
  return ctx
}
