'use client'

import { useContext } from 'react'
import { RunnerContext } from '@/providers/runner-provider'

export function useRunner() {
  const ctx = useContext(RunnerContext)
  if (!ctx) {
    throw new Error('useRunner must be used within a RunnerProvider')
  }
  return ctx
}
