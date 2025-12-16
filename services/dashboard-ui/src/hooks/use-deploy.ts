'use client'

import { useContext } from 'react'
import { DeployContext } from '@/providers/deploy-provider'

export function useDeploy() {
  const ctx = useContext(DeployContext)
  if (!ctx) {
    throw new Error('useDeploy must be used within a DeployProvider')
  }
  return ctx
}