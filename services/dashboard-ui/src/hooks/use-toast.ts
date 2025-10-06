'use client'

import { useContext } from 'react'
import { ToastContext } from '@/providers/toast-provider'

export function useToast() {
  const ctx = useContext(ToastContext)
  if (!ctx) {
    throw new Error('useToast must be used within an ToastProvider')
  }
  return ctx
}
