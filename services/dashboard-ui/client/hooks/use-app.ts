import { useContext } from 'react'
import { AppContext } from '@/providers/app-provider'
import type { TApp } from '@/types'

export function useApp(): { app: TApp; refresh: () => void } {
  const ctx = useContext(AppContext)
  if (!ctx) {
    throw new Error('useApp must be used within an AppProvider')
  }
  return ctx
}
