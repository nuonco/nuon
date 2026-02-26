'use client'

import { useContext } from 'react'
import { SandboxRunContext } from '@/providers/sandbox-run-provider'

export function useSandboxRun() {
  const ctx = useContext(SandboxRunContext)
  if (!ctx) {
    throw new Error('useSandboxRun must be used within a SandboxRunProvider')
  }
  return ctx
}