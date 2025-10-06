'use client'

import { useContext } from 'react'
import { AppContext } from '@/providers/app-provider'

export function useApp() {
  const ctx = useContext(AppContext)
  if (!ctx) {
    throw new Error('useApp must be used within an AppProvider')
  }
  return ctx
}
