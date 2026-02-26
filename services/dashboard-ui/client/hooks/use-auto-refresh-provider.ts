'use client'
import { useContext } from 'react'
import { AutoRefreshContext } from '@/providers/auto-refresh-provider'

export function useAutoRefresh() {
  const context = useContext(AutoRefreshContext)
  if (context === undefined) {
    throw new Error('useAutoRefresh must be used within an AutoRefreshProvider')
  }
  return context
}
