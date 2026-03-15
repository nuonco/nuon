import { useContext } from 'react'
import { SandboxBuildContext } from '@/providers/sandbox-build-provider'
import type { TAppSandboxBuild } from '@/types'

export function useSandboxBuild(): { build: TAppSandboxBuild } {
  const ctx = useContext(SandboxBuildContext)
  if (!ctx) {
    throw new Error('useSandboxBuild must be used within a SandboxBuildProvider')
  }
  return ctx
}
