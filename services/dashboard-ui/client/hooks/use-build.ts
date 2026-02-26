'use client'

import { useContext } from 'react'
import { BuildContext } from '@/providers/build-provider'

export function useBuild() {
  const ctx = useContext(BuildContext)
  if (!ctx) {
    throw new Error('useBuild must be used within a BuildProvider')
  }
  return ctx
}