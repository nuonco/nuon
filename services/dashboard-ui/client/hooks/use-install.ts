import { useContext } from 'react'
import { InstallContext } from '@/providers/install-provider'
import type { TInstall } from '@/types'

export function useInstall(): { install: TInstall; refresh: () => void } {
  const ctx = useContext(InstallContext)
  if (!ctx) {
    throw new Error('useInstall must be used within an InstallProvider')
  }
  return ctx
}
