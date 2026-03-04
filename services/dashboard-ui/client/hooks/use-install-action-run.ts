import { useContext } from 'react'
import { InstallActionRunContext } from '@/providers/install-action-run-provider'
import type { TInstallActionRun } from '@/types'

export function useInstallActionRun(): { installActionRun: TInstallActionRun; refresh: () => void } {
  const ctx = useContext(InstallActionRunContext)
  if (!ctx) {
    throw new Error(
      'useInstallActionRun must be used within an InstallActionRunProvider'
    )
  }
  return ctx
}
