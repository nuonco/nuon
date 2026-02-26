'use client'

import { useContext } from 'react'
import { InstallContext } from '@/providers/install-provider'

export function useInstall() {
  const ctx = useContext(InstallContext)
  if (!ctx) {
    throw new Error('useInstall must be used within an InstallProvider')
  }
  return ctx
}
