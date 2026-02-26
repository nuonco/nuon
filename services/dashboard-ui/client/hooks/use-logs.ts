'use client'

import { useContext } from 'react'
import { LogsContext } from '@/providers/logs-provider'

export function useLogs() {
  const ctx = useContext(LogsContext)
  if (!ctx) {
    throw new Error('useLogs must be used within an LogsProvider')
  }
  return ctx
}
