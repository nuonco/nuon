import { useContext } from 'react'
import { DeployContext } from '@/providers/deploy-provider'
import type { TDeploy } from '@/types'

export function useDeploy(): { deploy: TDeploy } {
  const ctx = useContext(DeployContext)
  if (!ctx) {
    throw new Error('useDeploy must be used within a DeployProvider')
  }
  return ctx
}