'use client'

import { useContext } from 'react'
import { InstallActionRunContext } from '@/providers/install-action-run-provider'

export function useInstallActionRun() {
  const ctx = useContext(InstallActionRunContext)
  if (!ctx) {
    throw new Error(
      'useInstallActionRun must be used within an InstallActionRunProvider'
    )
  }
  return ctx
}
