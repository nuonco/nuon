import { useContext } from 'react'
import { BuildContext } from '@/providers/build-provider'
import type { TBuild } from '@/types'

export function useBuild(): { build: TBuild } {
  const ctx = useContext(BuildContext)
  if (!ctx) {
    throw new Error('useBuild must be used within a BuildProvider')
  }
  return ctx
}