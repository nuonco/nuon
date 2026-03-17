import { useContext } from 'react'
import { BranchContext } from '@/providers/branch-provider'
import type { TAppBranch } from '@/types'

export function useBranch(): { branch: TAppBranch; refresh: () => void } {
  const ctx = useContext(BranchContext)
  if (!ctx) {
    throw new Error('useBranch must be used within a BranchProvider')
  }
  return ctx
}
