import { useContext } from 'react'
import { SandboxRunContext } from '@/providers/sandbox-run-provider'
import type { TSandboxRun } from '@/types'

export function useSandboxRun(): { sandboxRun: TSandboxRun } {
  const ctx = useContext(SandboxRunContext)
  if (!ctx) {
    throw new Error('useSandboxRun must be used within a SandboxRunProvider')
  }
  return ctx
}